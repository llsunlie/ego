package bootstrap

import (
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ego-server/internal/config"
	"ego-server/internal/platform/auth"
	"ego-server/internal/platform/metrics"
	"ego-server/internal/platform/ratelimit"

	"github.com/NYTimes/gziphandler"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "ego-server/proto/ego"
)

type Server struct {
	cfg         *config.Config
	grpcServer  *grpc.Server
	httpHandler http.Handler
	tlsConfig   *tls.Config
	certManager *autocert.Manager
	logger      *slog.Logger
}

func NewServer(cfg *config.Config, p *Platform, handler pb.EgoServer) *Server {
	var tlsConfig *tls.Config
	var certManager *autocert.Manager
	if cfg.TLSDomain != "" {
		certManager = &autocert.Manager{
			Cache:      autocert.DirCache("certs"),
			HostPolicy: autocert.HostWhitelist(cfg.TLSDomain),
			Prompt:     autocert.AcceptTOS,
		}
		tlsConfig = &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		}
		p.Logger.Info("TLS enabled", "domain", cfg.TLSDomain)
	}

	rateLimiter := ratelimit.New(cfg, p.Logger)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			auth.UnaryServerInterceptor(p.JWTKey, p.Logger),
			ratelimit.UnaryServerInterceptor(rateLimiter),
		),
	)
	pb.RegisterEgoServer(grpcServer, handler)
	reflection.Register(grpcServer)

	wrapped := grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(makeOriginChecker(cfg)),
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
	)

	// 静态文件目录（部署时提供前端 Web 文件，本地 dev 可留空走 flutter run）
	webDir := cfg.WebDir

	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ACME HTTP-01 challenge — Let's Encrypt domain validation.
		// Must be checked BEFORE any other routing so autocert can respond
		// to /.well-known/acme-challenge/ requests.
		if certManager != nil && strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			certManager.HTTPHandler(nil).ServeHTTP(w, r)
			return
		}

		// Health check — for UptimeRobot / load balancer probes.
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
			return
		}

		// Prometheus metrics endpoint.
		if r.URL.Path == "/metrics" {
			promhttp.Handler().ServeHTTP(w, r)
			return
		}

		if wrapped.IsGrpcWebRequest(r) || wrapped.IsAcceptableGrpcCorsRequest(r) {
			wrapped.ServeHTTP(w, r)
			return
		}
		// 非 gRPC 请求：提供前端静态文件
		if webDir != "" {
			http.FileServer(http.Dir(webDir)).ServeHTTP(w, r)
			return
		}
		if origin := r.Header.Get("Origin"); isOriginAllowed(origin, cfg) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with gzip compression for static assets (JS, CSS, HTML).
	h = gziphandler.GzipHandler(h)

	// Wrap with Prometheus HTTP metrics middleware.
	innerHandler := h
	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" || r.URL.Path == "/health" {
			innerHandler.ServeHTTP(w, r)
			return
		}

		metrics.HttpRequestsInFlight.Inc()
		defer metrics.HttpRequestsInFlight.Dec()

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		innerHandler.ServeHTTP(rec, r)

		elapsed := time.Since(start).Seconds()
		status := strconv.Itoa(rec.statusCode)
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		metrics.HttpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(elapsed)
	})

	return &Server{
		cfg:         cfg,
		grpcServer:  grpcServer,
		httpHandler: h,
		tlsConfig:   tlsConfig,
		certManager: certManager,
		logger:      p.Logger,
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// makeOriginChecker returns a grpcweb origin-check function that allows
// requests from configured origins and localhost (in dev mode).
func makeOriginChecker(cfg *config.Config) func(string) bool {
	return func(origin string) bool {
		return isOriginAllowed(origin, cfg)
	}
}

// isOriginAllowed returns true if the origin is allowed.
// Empty origin (same-origin request) is always allowed.
// In dev mode (TLS_DOMAIN empty), localhost origins are allowed.
// Otherwise the origin must be in the configured whitelist.
// An empty whitelist denies all cross-origin requests in production.
func isOriginAllowed(origin string, cfg *config.Config) bool {
	if origin == "" {
		return true
	}
	if cfg.TLSDomain == "" && isLocalhost(origin) {
		return true
	}
	for _, a := range cfg.AllowedOrigins() {
		if a == origin {
			return true
		}
	}
	return false
}

// isLocalhost returns true if the origin URL uses a localhost hostname.
func isLocalhost(origin string) bool {
	return strings.Contains(origin, "://localhost:") ||
		strings.Contains(origin, "://127.0.0.1:")
}

func (s *Server) Serve() error {
	// gRPC native port
	go func() {
		addr := ":" + s.cfg.GRPCPort
		var lis net.Listener
		var err error
		if s.tlsConfig != nil {
			lis, err = tls.Listen("tcp", addr, s.tlsConfig)
		} else {
			lis, err = net.Listen("tcp", addr)
		}
		if err != nil {
			s.logger.Error("gRPC listen failed", "error", err)
			panic(err)
		}
		s.logger.Info("gRPC server listening", "port", s.cfg.GRPCPort, "tls", s.tlsConfig != nil)
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Error("gRPC serve failed", "error", err)
			panic(err)
		}
	}()

	// Web plain port (always plain HTTP)
	go func() {
		addr := ":" + s.cfg.WebPort
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			s.logger.Error("web plain listen failed", "error", err)
			panic(err)
		}
		s.logger.Info("web plain server listening", "port", s.cfg.WebPort)
		if err := http.Serve(lis, s.httpHandler); err != nil {
			s.logger.Error("web plain serve failed", "error", err)
			panic(err)
		}
	}()

	// Web TLS port (TLS if enabled, else plain)
	addr := ":" + s.cfg.WebTLSPort
	var lis net.Listener
	var err error
	if s.tlsConfig != nil {
		lis, err = tls.Listen("tcp", addr, s.tlsConfig)
	} else {
		lis, err = net.Listen("tcp", addr)
	}
	if err != nil {
		return err
	}
	s.logger.Info("web TLS server listening", "port", s.cfg.WebTLSPort, "tls", s.tlsConfig != nil)
	return http.Serve(lis, s.httpHandler)
}
