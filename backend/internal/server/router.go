package server

import (
	"log/slog"
	"net/http"
)

func NewRouter(logger *slog.Logger) http.Handler {
	handler := NewHandler(logger)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handler.Root)
	mux.HandleFunc("GET /health/live", handler.Liveness)
	mux.HandleFunc("GET /health/ready", handler.Readiness)

	var router http.Handler = mux

	router = Logging(logger, router)
	router = Recovery(logger, router)
	router = RequestID(router)

	return router
}
