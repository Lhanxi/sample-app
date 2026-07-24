package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

	router := NewRouter(testLogger())

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
