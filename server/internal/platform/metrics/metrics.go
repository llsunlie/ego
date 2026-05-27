package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP requests
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ego_http_requests_total",
		Help: "Total HTTP requests.",
	}, []string{"method", "path", "status"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ego_http_request_duration_seconds",
		Help:    "HTTP request latency.",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
	}, []string{"method", "path"})

	HttpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ego_http_requests_in_flight",
		Help: "Current number of HTTP requests being served.",
	})

	// gRPC calls
	GrpcRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ego_grpc_requests_total",
		Help: "Total gRPC calls.",
	}, []string{"method", "status"})

	GrpcRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ego_grpc_request_duration_seconds",
		Help:    "gRPC call latency.",
		Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
	}, []string{"method"})

	// AI calls
	AiChatTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ego_ai_chat_total",
		Help: "Total AI chat completion calls.",
	}, []string{"model", "status"})

	AiChatDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ego_ai_chat_duration_seconds",
		Help:    "AI chat completion latency.",
		Buckets: []float64{.1, .5, 1, 2.5, 5, 10, 20, 30, 60},
	}, []string{"model"})

	AiChatTokens = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ego_ai_chat_tokens_total",
		Help: "Total tokens consumed by AI chat.",
	}, []string{"model", "type"}) // type=prompt|completion

	AiEmbedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ego_ai_embed_total",
		Help: "Total AI embedding calls.",
	}, []string{"model", "status"})

	AiEmbedDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ego_ai_embed_duration_seconds",
		Help:    "AI embedding latency.",
		Buckets: []float64{.1, .5, 1, 2.5, 5, 10, 20, 30},
	}, []string{"model"})

	AiEmbedTokens = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ego_ai_embed_tokens_total",
		Help: "Total tokens consumed by AI embeddings.",
	}, []string{"model"})

	// Active AI calls gauge
	AiCallsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ego_ai_calls_in_flight",
		Help: "Current number of AI calls in progress.",
	})
)
