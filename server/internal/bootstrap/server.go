package bootstrap

import (
	"log"
	"net"
	"net/http"

	"ego-server/internal/config"
	"ego-server/internal/platform/auth"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "ego-server/proto/ego"
)

type Server struct {
	cfg        *config.Config
	grpcServer *grpc.Server
	httpServer *http.Server
}

func NewServer(cfg *config.Config, p *Platform, handler pb.EgoServer) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(p.JWTKey)),
	)
	pb.RegisterEgoServer(grpcServer, handler)
	reflection.Register(grpcServer)

	wrapped := grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
	)

	httpServer := &http.Server{
		Addr: ":" + cfg.WebPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if wrapped.IsGrpcWebRequest(r) || wrapped.IsAcceptableGrpcCorsRequest(r) {
				wrapped.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusOK)
		}),
	}

	return &Server{
		cfg:        cfg,
		grpcServer: grpcServer,
		httpServer: httpServer,
	}
}

func (s *Server) Serve() error {
	// gRPC for native clients
	go func() {
		lis, err := net.Listen("tcp", ":"+s.cfg.Port)
		if err != nil {
			log.Fatalf("gRPC listen: %v", err)
		}
		log.Printf("gRPC server listening on :%s", s.cfg.Port)
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve: %v", err)
		}
	}()

	log.Printf("gRPC-web server listening on :%s", s.cfg.WebPort)
	return s.httpServer.ListenAndServe()
}
