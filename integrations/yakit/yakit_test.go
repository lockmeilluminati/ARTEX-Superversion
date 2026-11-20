package yakit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Autumn-27/artex/integrations"
)

func TestHealthWithMockEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"mock-1.0","capabilities":["health"]}`))
	}))
	defer server.Close()

	got := New(Config{HealthURL: server.URL}).CheckHealth(context.Background())
	if got.State != integrations.StateAvailable {
		t.Fatalf("state = %q, want available (diagnostic: %s)", got.State, got.Diagnostic)
	}
	if got.Version != "mock-1.0" {
		t.Fatalf("version = %q, want mock-1.0", got.Version)
	}
	if len(got.Capabilities) != 1 || got.Capabilities[0] != "health" {
		t.Fatalf("unexpected capabilities: %#v", got.Capabilities)
	}
}

func TestRejectsNonLoopbackEndpoint(t *testing.T) {
	got := New(Config{HealthURL: "https://example.com/health"}).CheckHealth(context.Background())
	if got.State != integrations.StateMisconfigured {
		t.Fatalf("state = %q, want misconfigured", got.State)
	}
}
