// Package toolregistry defines metadata contracts for tools known to ARTEX.
// Persistence and UI integration are intentionally left for later phases.
package toolregistry

import "time"

type Lifecycle string
const (
	LifecycleCandidate Lifecycle = "candidate"
	LifecycleAvailable Lifecycle = "available"
	LifecycleInstalled Lifecycle = "installed"
	LifecycleDeprecated Lifecycle = "deprecated"
	LifecycleQuarantined Lifecycle = "quarantined"
)

type Permission string
const (
	PermissionDiscover Permission = "discover"
	PermissionReadOnly Permission = "read_only"
	PermissionExecute Permission = "execute"
	PermissionInstall Permission = "install"
)

type Source struct {
	Kind string `json:"kind"`
	URL string `json:"url,omitempty"`
	Revision string `json:"revision,omitempty"`
	ObservedAt time.Time `json:"observed_at,omitempty"`
}

type Dependency struct {
	Name string `json:"name"`
	Constraint string `json:"constraint,omitempty"`
	Kind string `json:"kind,omitempty"`
}

type Manifest struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Description string `json:"description,omitempty"`
	Version string `json:"version,omitempty"`
	License string `json:"license,omitempty"`
	Maintainer string `json:"maintainer,omitempty"`
	CanonicalRepository string `json:"canonical_repository,omitempty"`
	Sources []Source `json:"sources,omitempty"`
	Platforms []string `json:"platforms,omitempty"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
	Lifecycle Lifecycle `json:"lifecycle"`
	Verified bool `json:"verified"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (m Manifest) Validate() error {
	if m.ID == "" { return ErrMissingID }
	if m.Name == "" { return ErrMissingName }
	if m.Lifecycle == "" { return ErrMissingLifecycle }
	return nil
}
