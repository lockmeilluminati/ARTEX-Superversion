package toolregistry

import (
	"errors"
	"sync"
)

var (
	ErrMissingID = errors.New("tool manifest ID is required")
	ErrMissingName = errors.New("tool manifest name is required")
	ErrMissingLifecycle = errors.New("tool lifecycle is required")
	ErrDuplicate = errors.New("tool manifest already exists")
	ErrNotFound = errors.New("tool manifest not found")
	ErrNilDB = errors.New("tool registry database is nil")
)

type Registry struct {
	mu sync.RWMutex
	items map[string]Manifest
}

func New() *Registry { return &Registry{items: make(map[string]Manifest)} }

func (r *Registry) Register(m Manifest) error {
	if err := m.Validate(); err != nil { return err }
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[m.ID]; ok { return ErrDuplicate }
	r.items[m.ID] = m
	return nil
}

func (r *Registry) Get(id string) (Manifest, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.items[id]
	return m, ok
}

func (r *Registry) List() []Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Manifest, 0, len(r.items))
	for _, m := range r.items { out = append(out, m) }
	return out
}
