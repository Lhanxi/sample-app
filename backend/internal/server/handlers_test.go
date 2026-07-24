package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRootHandler(t *testing.T) {
	logger := testLogger()
	handler := NewHandler(logger)

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
	handler := NewHandler(logger)

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
	logger := testLogger()
	handler := NewHandler(logger)

	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	recorder := httptest.NewRecorder()

	handler.Readiness(recorder, request)

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

	if body["status"] != "ready" {
		t.Errorf(
			"status = %q; want %q",
			body["status"],
			"ready",
		)
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
