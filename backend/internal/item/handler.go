package item

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type ItemService interface {
	List(ctx context.Context) ([]Item, error)
	GetByID(ctx context.Context, id string) (Item, error)
	Create(ctx context.Context, input CreateItemRequest) (Item, error)
	Update(ctx context.Context, id string, input UpdateItemRequest) (Item, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	service ItemService
	logger  *slog.Logger
}

func NewHandler(service ItemService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := validPathID(w, r)
	if !ok {
		return
	}

	foundItem, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, foundItem)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateItemRequest
	if err := decodeJSON(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	createdItem, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createdItem)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := validPathID(w, r)
	if !ok {
		return
	}

	var input UpdateItemRequest
	if err := decodeJSON(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	updatedItem, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updatedItem)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := validPathID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeErrorJSON(w, http.StatusBadRequest, ErrInvalidInput.Error())
	case errors.Is(err, ErrNotFound):
		writeErrorJSON(w, http.StatusNotFound, ErrNotFound.Error())
	case errors.Is(err, ErrConflict):
		writeErrorJSON(w, http.StatusConflict, ErrConflict.Error())
	default:
		h.logger.Error("item request failed", "error", err)
		writeErrorJSON(w, http.StatusInternalServerError, "internal server error")
	}
}

func validPathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid item ID")
		return "", false
	}

	return id, true
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}

	return nil
}

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to encode item response", "error", err)
	}
}
