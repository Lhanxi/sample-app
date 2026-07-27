package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeDatabase struct {
	err error
}

func (f fakeDatabase) Ping(context.Context) error {
	return f.err
}

func TestRootHandler(t *testing.T) {
	logger := testLogger()
	handler := NewHandler(logger, fakeDatabase{})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	handler.Root(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d; want %d",
			response.StatusCode,
			http.StatusOK,
		)
	}

	if contentType := response.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf(
			"Content-Type = %q; want %q",
			contentType,
			"application/json",
		)
	}

	var body map[string]string

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["service"] != "sample-backend" {
		t.Errorf(
			"service = %q; want %q",
			body["service"],
			"sample-backend",
		)
	}

	if body["message"] != "sample backend is running" {
		t.Errorf(
			"message = %q; want %q",
			body["message"],
			"sample backend is running",
		)
	}
}

func TestLivenessHandler(t *testing.T) {
	logger := testLogger()
	handler := NewHandler(logger, fakeDatabase{})

	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	recorder := httptest.NewRecorder()

	handler.Liveness(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d; want %d",
			response.StatusCode,
			http.StatusOK,
		)
	}

	var body map[string]string

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["status"] != "alive" {
		t.Errorf(
			"status = %q; want %q",
			body["status"],
			"alive",
		)
	}
}

func TestReadinessHandler(t *testing.T) {
	tests := []struct {
		name           string
		databaseError  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "ready database",
			expectedStatus: http.StatusOK,
			expectedBody:   "ready",
		},
		{
			name:           "unavailable database",
			databaseError:  errors.New("database unavailable"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "not ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(testLogger(), fakeDatabase{err: tt.databaseError})
			request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
			recorder := httptest.NewRecorder()

			handler.Readiness(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != tt.expectedStatus {
				t.Fatalf(
					"status code = %d; want %d",
					response.StatusCode,
					tt.expectedStatus,
				)
			}

			var body map[string]string

			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			if body["status"] != tt.expectedBody {
				t.Errorf(
					"status = %q; want %q",
					body["status"],
					tt.expectedBody,
				)
			}
		})
	}
}

func testLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(
			io.Discard,
			nil,
		),
	)
}
