package main

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"ego-server/internal/auth"
	"ego-server/internal/config"
	"ego-server/internal/db"
	"ego-server/internal/db/sqlc"
	"ego-server/internal/login"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "ego-server/proto/ego"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	jwtKey := []byte(cfg.JWTSecret)
	expHours, err := strconv.Atoi(cfg.JWTExpHours)
	if err != nil {
		log.Fatalf("invalid JWT_EXP_HOURS: %v", err)
	}
	jwtExp := time.Duration(expHours) * time.Hour

	server := grpc.NewServer(
		grpc.UnaryInterceptor(auth.UnaryServerInterceptor(jwtKey)),
	)

	loginHandler := login.NewHandler(sqlc.New(pool), jwtKey, jwtExp)
	pb.RegisterEgoServer(server, loginHandler)
	reflection.Register(server)

	// gRPC for native clients
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.Port)
		if err != nil {
			log.Fatalf("gRPC listen: %v", err)
		}
		log.Printf("gRPC server listening on :%s", cfg.Port)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("gRPC serve: %v", err)
		}
	}()

	// gRPC-web for browser clients
	wrapped := grpcweb.WrapServer(server,
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
	)

	httpSrv := &http.Server{
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

	log.Printf("gRPC-web server listening on :%s", cfg.WebPort)
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Fatalf("gRPC-web serve: %v", err)
	}
}
