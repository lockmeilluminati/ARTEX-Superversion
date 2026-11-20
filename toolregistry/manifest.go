// Package toolregistry defines the unified metadata contract for tools known to ARTEX.
// Execution remains owned by the existing ARTEX handlers and security gates.
package toolregistry

import (
	"encoding/json"
	"time"
)

type Lifecycle string

const (
	LifecycleCandidate    Lifecycle = "candidate"
	LifecycleAvailable    Lifecycle = "available"
	LifecycleInstalled    Lifecycle = "installed"
	LifecycleDeprecated   Lifecycle = "deprecated"
	LifecycleQuarantined  Lifecycle = "quarantined"
)

type Availability string

const (
	AvailabilityUnknown    Availability = "unknown"
	AvailabilityRegistered Availability = "registered"
	AvailabilityPresent    Availability = "present"
	AvailabilityConnected  Availability = "connected"
)

type TrustState string

const (
	TrustUnknown            TrustState = "unknown"
	TrustCodeDefined        TrustState = "code_defined"
	TrustOperatorConfigured TrustState = "operator_configured"
	TrustExternallyVerified TrustState = "externally_verified"
)

type Permission string

const (
	PermissionDiscover Permission = "discover"
	PermissionReadOnly Permission = "read_only"
	PermissionExecute  Permission = "execute"
	PermissionInstall  Permission = "install"
)

type Source struct {
	Kind       string    `json:"kind"`
	URL        string    `json:"url,omitempty"`
	Revision   string    `json:"revision,omitempty"`
	ObservedAt time.Time `json:"observed_at,omitempty"`
}

type Dependency struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint,omitempty"`
	Kind       string `json:"kind,omitempty"`
}

// ExecutionMode describes the interaction surface an agent needs to operate a tool.
type ExecutionMode string

const (
	ExecutionUnknown ExecutionMode = "unknown"
	ExecutionCLI     ExecutionMode = "cli"
	ExecutionGUI     ExecutionMode = "gui"
	ExecutionAPI     ExecutionMode = "api"
	ExecutionMCP     ExecutionMode = "mcp"
	ExecutionHybrid  ExecutionMode = "hybrid"
)

// VerificationStatus records interface verification independently from metadata review.
type VerificationStatus string

const (
	VerificationUnverified VerificationStatus = "unverified"
	VerificationPartial    VerificationStatus = "partial"
	VerificationVerified   VerificationStatus = "verified"
	VerificationFailed    VerificationStatus = "failed"
)

// InstallationStatus is scoped to the environment in which discovery was performed.
type InstallationStatus string

const (
	InstallationUnknown      InstallationStatus = "unknown"
	InstallationInstalled    InstallationStatus = "installed"
	InstallationMissing      InstallationStatus = "missing"
	InstallationNotApplicable InstallationStatus = "not_applicable"
)

// AgentCapability is a machine-readable requirement that an agent must satisfy.
type AgentCapability string

const (
	CapabilityTerminal          AgentCapability = "terminal"
	CapabilityScreenCapture     AgentCapability = "screen_capture"
	CapabilityVisualInteraction AgentCapability = "visual_interaction"
	CapabilityBrowser           AgentCapability = "browser"
	CapabilityAPI               AgentCapability = "api"
	CapabilityMCP               AgentCapability = "mcp"
)

// PerformanceProfile contains measurements, not a guessed speed label.
type PerformanceProfile struct {
	SampleCount     int        `json:"sample_count"`
	MedianStartupMS int64      `json:"median_startup_ms,omitempty"`
	MedianRunMS     int64      `json:"median_run_ms,omitempty"`
	LastMeasuredAt  *time.Time `json:"last_measured_at,omitempty"`
}

type Manifest struct {
	ID                         string             `json:"id"`
	LegacyKey                  string             `json:"legacy_key,omitempty"`
	Name                       string             `json:"name"`
	Description                string             `json:"description,omitempty"`
	Category                   string             `json:"category,omitempty"`
	Version                    string             `json:"version,omitempty"`
	License                    string             `json:"license,omitempty"`
	Maintainer                 string             `json:"maintainer,omitempty"`
	CanonicalRepository        string             `json:"canonical_repository,omitempty"`
	Sources                    []Source           `json:"sources,omitempty"`
	Platforms                  []string           `json:"platforms,omitempty"`
	Dependencies               []Dependency       `json:"dependencies,omitempty"`
	Capabilities               []string           `json:"capabilities,omitempty"`
	Permissions                []Permission       `json:"permissions,omitempty"`
	AgentBindings              []string           `json:"agent_bindings,omitempty"`
	ExecutionKind              string             `json:"execution_kind,omitempty"`
	ExecutionMode              ExecutionMode      `json:"execution_mode,omitempty"`
	VerificationStatus         VerificationStatus `json:"verification_status,omitempty"`
	InstallationStatus         InstallationStatus `json:"installation_status,omitempty"`
	VerificationVersion        string             `json:"verification_version,omitempty"`
	VerifiedAt                 *time.Time         `json:"verified_at,omitempty"`
	RequiredAgentCapabilities  []AgentCapability  `json:"required_agent_capabilities,omitempty"`
	Performance                *PerformanceProfile `json:"performance,omitempty"`
	InputSchema                json.RawMessage    `json:"input_schema,omitempty"`
	Enabled                    bool               `json:"enabled"`
	Availability               Availability       `json:"availability"`
	Trust                      TrustState         `json:"trust"`
	ApprovalRequired           bool               `json:"approval_required"`
	Lifecycle                  Lifecycle          `json:"lifecycle"`
	Verified                   bool               `json:"verified"`
	CreatedAt                  time.Time          `json:"created_at,omitempty"`
	UpdatedAt                  time.Time          `json:"updated_at,omitempty"`
}

func (m Manifest) Validate() error {
	if m.ID == "" {
		return ErrMissingID
	}
	if m.Name == "" {
		return ErrMissingName
	}
	if m.Lifecycle == "" {
		return ErrMissingLifecycle
	}
	return nil
}
