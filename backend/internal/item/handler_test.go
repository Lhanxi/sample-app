package item

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testItemID = "b8f574d6-1d0d-4f63-b4a8-4ec847bd9f1d"

type fakeItemService struct {
	items       []Item
	item        Item
	err         error
	createInput CreateItemRequest
	updateInput UpdateItemRequest
	deletedID   string
}

func (f *fakeItemService) List(context.Context) ([]Item, error) {
	return f.items, f.err
}

func (f *fakeItemService) GetByID(context.Context, string) (Item, error) {
	return f.item, f.err
}

func (f *fakeItemService) Create(
	_ context.Context,
	input CreateItemRequest,
) (Item, error) {
	f.createInput = input
	return f.item, f.err
}

func (f *fakeItemService) Update(
	_ context.Context,
	_ string,
	input UpdateItemRequest,
) (Item, error) {
	f.updateInput = input
	return f.item, f.err
}

func (f *fakeItemService) Delete(_ context.Context, id string) error {
	f.deletedID = id
	return f.err
}

func TestHandlerList(t *testing.T) {
	service := &fakeItemService{
		items: []Item{{ID: testItemID, Name: "Desk"}},
	}
	response := serveItemRequest(t, service, http.MethodGet, "/api/v1/items", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", response.Code, http.StatusOK)
	}

	var body []Item
	decodeResponse(t, response, &body)
	if len(body) != 1 || body[0].Name != "Desk" {
		t.Errorf("body = %#v; want one Desk item", body)
	}
}

func TestHandlerGet(t *testing.T) {
	t.Run("retrieves item", func(t *testing.T) {
		service := &fakeItemService{item: Item{ID: testItemID, Name: "Desk"}}
		response := serveItemRequest(
			t,
			service,
			http.MethodGet,
			"/api/v1/items/"+testItemID,
			nil,
		)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d; want %d", response.Code, http.StatusOK)
		}
	})

	t.Run("rejects invalid UUID", func(t *testing.T) {
		response := serveItemRequest(
			t,
			&fakeItemService{},
			http.MethodGet,
			"/api/v1/items/not-a-uuid",
			nil,
		)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; want %d", response.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns missing item", func(t *testing.T) {
		response := serveItemRequest(
			t,
			&fakeItemService{err: ErrNotFound},
			http.MethodGet,
			"/api/v1/items/"+testItemID,
			nil,
		)

		if response.Code != http.StatusNotFound {
			t.Fatalf("status = %d; want %d", response.Code, http.StatusNotFound)
		}
	})
}

func TestHandlerCreate(t *testing.T) {
	t.Run("creates valid item", func(t *testing.T) {
		service := &fakeItemService{item: Item{ID: testItemID, Name: "Desk"}}
		response := serveItemRequest(
			t,
			service,
			http.MethodPost,
			"/api/v1/items",
			[]byte(`{"name":"Desk","description":"Standing desk"}`),
		)

		if response.Code != http.StatusCreated {
			t.Fatalf("status = %d; want %d", response.Code, http.StatusCreated)
		}
		if service.createInput.Name != "Desk" {
			t.Errorf("create name = %q; want %q", service.createInput.Name, "Desk")
		}
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		response := serveItemRequest(
			t,
			&fakeItemService{},
			http.MethodPost,
			"/api/v1/items",
			[]byte(`{"name":`),
		)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; want %d", response.Code, http.StatusBadRequest)
		}
	})

	t.Run("rejects blank name", func(t *testing.T) {
		response := serveItemRequest(
			t,
			&fakeItemService{err: ErrInvalidInput},
			http.MethodPost,
			"/api/v1/items",
			[]byte(`{"name":" "}`),
		)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; want %d", response.Code, http.StatusBadRequest)
		}
	})
}

func TestHandlerUpdate(t *testing.T) {
	service := &fakeItemService{item: Item{ID: testItemID, Name: "Chair"}}
	response := serveItemRequest(
		t,
		service,
		http.MethodPut,
		"/api/v1/items/"+testItemID,
		[]byte(`{"name":"Chair","description":"Office chair"}`),
	)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", response.Code, http.StatusOK)
	}
	if service.updateInput.Name != "Chair" {
		t.Errorf("update name = %q; want %q", service.updateInput.Name, "Chair")
	}
}

func TestHandlerDelete(t *testing.T) {
	service := &fakeItemService{}
	response := serveItemRequest(
		t,
		service,
		http.MethodDelete,
		"/api/v1/items/"+testItemID,
		nil,
	)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want %d", response.Code, http.StatusNoContent)
	}
	if service.deletedID != testItemID {
		t.Errorf("deleted ID = %q; want %q", service.deletedID, testItemID)
	}
}

func TestHandlerUnexpectedError(t *testing.T) {
	response := serveItemRequest(
		t,
		&fakeItemService{err: errors.New("database failed")},
		http.MethodGet,
		"/api/v1/items",
		nil,
	)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d; want %d", response.Code, http.StatusInternalServerError)
	}
}

func TestHandlerConflict(t *testing.T) {
	response := serveItemRequest(
		t,
		&fakeItemService{err: ErrConflict},
		http.MethodPost,
		"/api/v1/items",
		[]byte(`{"name":"Desk"}`),
	)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d; want %d", response.Code, http.StatusConflict)
	}
}

func serveItemRequest(
	t *testing.T,
	service ItemService,
	method string,
	path string,
	body []byte,
) *httptest.ResponseRecorder {
	t.Helper()

	handler := NewHandler(
		service,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/items", handler.List)
	mux.HandleFunc("POST /api/v1/items", handler.Create)
	mux.HandleFunc("GET /api/v1/items/{id}", handler.Get)
	mux.HandleFunc("PUT /api/v1/items/{id}", handler.Update)
	mux.HandleFunc("DELETE /api/v1/items/{id}", handler.Delete)

	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, destination any) {
	t.Helper()

	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
