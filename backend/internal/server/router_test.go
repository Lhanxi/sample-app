package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lhanxi/sample-app/backend/internal/item"
)

type routerItemService struct{}

func (routerItemService) List(context.Context) ([]item.Item, error) {
	return []item.Item{}, nil
}

func (routerItemService) GetByID(context.Context, string) (item.Item, error) {
	return item.Item{}, item.ErrNotFound
}

func (routerItemService) Create(
	context.Context,
	item.CreateItemRequest,
) (item.Item, error) {
	return item.Item{}, nil
}

func (routerItemService) Update(
	context.Context,
	string,
	item.UpdateItemRequest,
) (item.Item, error) {
	return item.Item{}, nil
}

func (routerItemService) Delete(context.Context, string) error {
	return nil
}

func TestRouterRoutes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "root route",
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]string{
				"service": "sample-backend",
				"message": "sample backend is running",
			},
		},
		{
			name:           "liveness route",
			method:         http.MethodGet,
			path:           "/health/live",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]string{
				"status": "alive",
			},
		},
		{
			name:           "readiness route",
			method:         http.MethodGet,
			path:           "/health/ready",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]string{
				"status": "ready",
			},
		},
		{
			name:           "unknown route",
			method:         http.MethodGet,
			path:           "/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   nil,
		},
		{
			name:           "unsupported method",
			method:         http.MethodPost,
			path:           "/health/live",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
	}

	itemHandler := item.NewHandler(routerItemService{}, testLogger())
	router := NewRouter(testLogger(), fakeDatabase{}, itemHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != tt.expectedStatus {
				t.Fatalf(
					"status code = %d; want %d",
					response.StatusCode,
					tt.expectedStatus,
				)
			}

			if tt.expectedBody == nil {
				return
			}

			var body map[string]string

			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			for key, expectedValue := range tt.expectedBody {
				if actualValue := body[key]; actualValue != expectedValue {
					t.Errorf(
						"body[%q] = %q; want %q",
						key,
						actualValue,
						expectedValue,
					)
				}
			}
		})
	}
}

func TestRouterRegistersItemRoutes(t *testing.T) {
	itemHandler := item.NewHandler(routerItemService{}, testLogger())
	router := NewRouter(testLogger(), fakeDatabase{}, itemHandler)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "list items",
			method:         http.MethodGet,
			path:           "/api/v1/items",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "create item",
			method:         http.MethodPost,
			path:           "/api/v1/items",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "get item",
			method:         http.MethodGet,
			path:           "/api/v1/items/not-a-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "update item",
			method:         http.MethodPut,
			path:           "/api/v1/items/not-a-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "delete item",
			method:         http.MethodDelete,
			path:           "/api/v1/items/b8f574d6-1d0d-4f63-b4a8-4ec847bd9f1d",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.expectedStatus {
				t.Errorf(
					"status code = %d; want %d",
					recorder.Code,
					tt.expectedStatus,
				)
			}
		})
	}
}
