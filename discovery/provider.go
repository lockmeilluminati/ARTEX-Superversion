// Package discovery collects candidate metadata only. Discovery never installs or executes tools.
package discovery

import "context"

type Query struct {
	Keywords []string
	Platforms []string
	Limit int
}

type Candidate struct {
	ID string
	Name string
	Description string
	RepositoryURI string
	SourceKind string
	ObservedAt string
	Signals []string
}

type Provider interface {
	Name() string
	Search(context.Context, Query) ([]Candidate, error)
}

type Service struct { providers []Provider }
func New(providers ...Provider) *Service { return &Service{providers: providers} }
func (s *Service) Providers() []string {
	names := make([]string, 0, len(s.providers))
	for _, p := range s.providers { names = append(names, p.Name()) }
	return names
}
