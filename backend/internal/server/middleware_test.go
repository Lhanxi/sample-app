package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDUsesExistingHeader(t *testing.T) {
	const expectedRequestID = "existing-request-id"

	handler := RequestID(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := requestIDFromContext(r.Context())

			if requestID != expectedRequestID {
				t.Errorf(
					"request ID in context = %q; want %q",
					requestID,
					expectedRequestID,
				)
			}

			w.WriteHeader(http.StatusOK)
		}),
	)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", expectedRequestID)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if actual := recorder.Header().Get("X-Request-ID"); actual != expectedRequestID {
		t.Errorf(
			"X-Request-ID response header = %q; want %q",
			actual,
			expectedRequestID,
		)
	}
}

func TestRequestIDGeneratesHeader(t *testing.T) {
	handler := RequestID(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := requestIDFromContext(r.Context())

			if requestID == "" {
				t.Error("request ID in context is empty")
			}

			w.WriteHeader(http.StatusOK)
		}),
	)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	requestID := recorder.Header().Get("X-Request-ID")

	if requestID == "" {
		t.Fatal("X-Request-ID response header is empty")
	}

	if len(requestID) != 32 {
		t.Errorf(
			"request ID length = %d; want %d",
			len(requestID),
			32,
		)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var logOutput bytes.Buffer

	logger := slog.New(
		slog.NewJSONHandler(
			&logOutput,
			nil,
		),
	)

	handler := RequestID(
		Logging(
			logger,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)

				if _, err := w.Write([]byte(`{"status":"created"}`)); err != nil {
					t.Fatalf("failed to write response: %v", err)
				}
			}),
		),
	)

	request := httptest.NewRequest(http.MethodPost, "/items", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	logged := logOutput.String()

	expectedValues := []string{
		`"msg":"http request"`,
		`"method":"POST"`,
		`"path":"/items"`,
		`"status":201`,
	}

	for _, expected := range expectedValues {
		if !strings.Contains(logged, expected) {
			t.Errorf(
				"log output %q does not contain %q",
				logged,
				expected,
			)
		}
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	var logOutput bytes.Buffer

	logger := slog.New(
		slog.NewJSONHandler(
			&logOutput,
			nil,
		),
	)

	handler := RequestID(
		Recovery(
			logger,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("unexpected failure")
			}),
		),
	)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf(
			"status code = %d; want %d",
			response.StatusCode,
			http.StatusInternalServerError,
		)
	}

	if !strings.Contains(logOutput.String(), "panic recovered") {
		t.Errorf(
			"log output %q does not contain panic recovery message",
			logOutput.String(),
		)
	}
}
