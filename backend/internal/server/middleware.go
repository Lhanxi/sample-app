package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const requestIDKey contextKey = "request-id"

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func CORS(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		originAllowed := origin != "" && origin == allowedOrigin

		w.Header().Add("Vary", "Origin")

		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, DELETE, OPTIONS",
			)
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, X-Request-ID",
			)
		}

		if r.Method == http.MethodOptions {
			if !originAllowed {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	n, err := r.ResponseWriter.Write(body)
	r.bytes += n
	return n, err
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}

		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		attributes := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"bytes", recorder.bytes,
			"duration", time.Since(start),
			"request_id", requestIDFromContext(r.Context()),
			"remote_address", r.RemoteAddr,
		}
		attributes = append(attributes, traceLogAttributes(r.Context())...)

		logger.Info("http request", attributes...)
	})
}

func Recovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				attributes := []any{
					"error", recovered,
					"stack", string(debug.Stack()),
					"request_id", requestIDFromContext(r.Context()),
				}
				attributes = append(attributes, traceLogAttributes(r.Context())...)

				logger.Error("panic recovered", attributes...)

				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func Tracing(next http.Handler) http.Handler {
	instrumented := otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			if r.Pattern != "" {
				span := trace.SpanFromContext(r.Context())
				span.SetName(r.Method + " " + r.Pattern)
				span.SetAttributes(attribute.String("http.route", r.Pattern))
			}
		}),
		"http.request",
		otelhttp.WithFilter(func(r *http.Request) bool {
			switch r.URL.Path {
			case "/metrics", "/health/live", "/health/ready":
				return false
			default:
				return true
			}
		}),
	)

	return instrumented
}

func newRequestID() string {
	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err != nil {
		return "unknown"
	}

	return hex.EncodeToString(bytes)
}

func requestIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}

	return value
}

func traceLogAttributes(ctx context.Context) []any {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return nil
	}

	return []any{
		"trace_id", spanContext.TraceID().String(),
		"span_id", spanContext.SpanID().String(),
	}
}
