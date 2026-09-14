package telemetry

import (
	"context"
	"testing"
)

func TestNewTracerProviderWithOTLPEndpoint(t *testing.T) {
	shutdown, err := NewTracerProvider(
		context.Background(),
		"sample-backend",
		"test",
		"127.0.0.1:4317",
	)
	if err != nil {
		t.Fatalf("NewTracerProvider() returned an unexpected error: %v", err)
	}

	shutdownContext, cancel := context.WithCancel(context.Background())
	cancel()

	_ = shutdown(shutdownContext)
}
