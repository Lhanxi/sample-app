package server

import (
	"log/slog"
	"net/http"

	"github.com/Lhanxi/sample-app/backend/internal/item"
)

func NewRouter(
	logger *slog.Logger,
	db DatabasePinger,
	itemHandler *item.Handler,
) http.Handler {
	handler := NewHandler(logger, db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handler.Root)
	mux.HandleFunc("GET /health/live", handler.Liveness)
	mux.HandleFunc("GET /health/ready", handler.Readiness)

	mux.HandleFunc("GET /api/v1/items", itemHandler.List)
	mux.HandleFunc("POST /api/v1/items", itemHandler.Create)
	mux.HandleFunc("GET /api/v1/items/{id}", itemHandler.Get)
	mux.HandleFunc("PUT /api/v1/items/{id}", itemHandler.Update)
	mux.HandleFunc("DELETE /api/v1/items/{id}", itemHandler.Delete)

	var router http.Handler = mux

	router = Logging(logger, router)
	router = Recovery(logger, router)
	router = RequestID(router)

	return router
}
