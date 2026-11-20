// Package integrations provides a small, read-only integration foundation.
// It deliberately does not execute security tools or start scans.
package integrations

import (
	"context"
	"errors"
	"sync"
	"time"
)

type HealthState string

const (
	StateUnknown       HealthState = "unknown"
	StateAvailable     HealthState = "available"
	StateUnavailable   HealthState = "unavailable"
	StateMisconfigured HealthState = "misconfigured"
)

type Health struct {
	IntegrationID string      `json:"integration_id"`
	Name          string      `json:"name"`
	State         HealthState `json:"state"`
	Version       string      `json:"version,omitempty"`
	Capabilities  []string    `json:"capabilities,omitempty"`
	CheckedAt     time.Time   `json:"checked_at"`
	Diagnostic    string      `json:"diagnostic,omitempty"`
}

// Adapter separates connectivity checks from tool execution.
type Adapter interface {
	ID() string
	Name() string
	CheckHealth(context.Context) Health
}

var ErrDuplicateAdapter = errors.New("integration adapter already registered")

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[string]Adapter)}
}

func (r *Registry) Register(adapter Adapter) error {
	if adapter == nil || adapter.ID() == "" {
		return errors.New("integration adapter and ID are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.adapters[adapter.ID()]; exists {
		return ErrDuplicateAdapter
	}
	r.adapters[adapter.ID()] = adapter
	return nil
}

func (r *Registry) Get(id string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[id]
	return a, ok
}

func (r *Registry) Health(ctx context.Context) []Health {
	r.mu.RLock()
	adapters := make([]Adapter, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		adapters = append(adapters, adapter)
	}
	r.mu.RUnlock()

	out := make([]Health, 0, len(adapters))
	for _, adapter := range adapters {
		out = append(out, adapter.CheckHealth(ctx))
	}
	return out
}
