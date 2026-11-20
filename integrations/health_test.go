package integrations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPProbeHealthy(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"version":"test"}`))
	}))
	defer s.Close()
	got := (HTTPProbe{Endpoint: s.URL}).Check(context.Background(), "mock", "Mock")
	if got.State != StateAvailable || got.Version != "test" {
		t.Fatalf("unexpected health: %#v", got)
	}
}

func TestHTTPProbeDoesNotFollowRedirect(t *testing.T) {
	targetCalled := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		targetCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer redirect.Close()

	got := (HTTPProbe{Endpoint: redirect.URL}).Check(context.Background(), "mock", "Mock")
	if got.State != StateUnavailable {
		t.Fatalf("state = %q, want unavailable", got.State)
	}
	if targetCalled {
		t.Fatal("probe followed redirect")
	}
}

type testAdapter struct{}
func (testAdapter) ID() string { return "test" }
func (testAdapter) Name() string { return "Test" }
func (testAdapter) CheckHealth(context.Context) Health {
	return Health{IntegrationID: "test", Name: "Test", State: StateAvailable}
}

func TestRegistryRejectsDuplicate(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(testAdapter{}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(testAdapter{}); err != ErrDuplicateAdapter {
		t.Fatalf("error = %v, want duplicate adapter", err)
	}
	if got := r.Health(context.Background()); len(got) != 1 || got[0].State != StateAvailable {
		t.Fatalf("unexpected registry health: %#v", got)
	}
}
