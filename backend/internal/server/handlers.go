package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type DatabasePinger interface {
	Ping(context.Context) error
}

type Handler struct {
	logger *slog.Logger
	db     DatabasePinger
}

func NewHandler(logger *slog.Logger, db DatabasePinger) *Handler {
	return &Handler{
		logger: logger,
		db:     db,
	}
}

func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "sample-backend",
		"message": "sample backend is running",
	})
}

func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "alive",
	})
}

func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(r.Context()); err != nil {
		h.logger.Error("database readiness check failed", "error", err)

		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
		})

		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
