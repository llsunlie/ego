package bootstrap

import (
	"log/slog"
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
	logger     *slog.Logger
}

func NewServer(cfg *config.Config, p *Platform, handler pb.EgoServer) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(p.JWTKey, p.Logger)),
	)
	pb.RegisterEgoServer(grpcServer, handler)
	reflection.Register(grpcServer)

	wrapped := grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
	)

	// 静态文件目录（部署时提供前端 Web 文件，本地 dev 可留空走 flutter run）
	webDir := cfg.WebDir

	httpServer := &http.Server{
		Addr: ":" + cfg.WebPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if wrapped.IsGrpcWebRequest(r) || wrapped.IsAcceptableGrpcCorsRequest(r) {
				wrapped.ServeHTTP(w, r)
				return
			}
			// 非 gRPC 请求：提供前端静态文件
			if webDir != "" {
				http.FileServer(http.Dir(webDir)).ServeHTTP(w, r)
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
		logger:     p.Logger,
	}
}

func (s *Server) Serve() error {
	go func() {
		lis, err := net.Listen("tcp", ":"+s.cfg.Port)
		if err != nil {
			s.logger.Error("gRPC listen failed", "error", err)
			panic(err)
		}
		s.logger.Info("gRPC server listening", "port", s.cfg.Port)
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Error("gRPC serve failed", "error", err)
			panic(err)
		}
	}()

	s.logger.Info("gRPC-web server listening", "port", s.cfg.WebPort)
	return s.httpServer.ListenAndServe()
}
