// Package executionprofile describes future execution constraints only.
// It does not start processes or replace ARTEX intercept/guard controls.
package executionprofile

import "time"

type Isolation string
const (
	IsolationNone Isolation = "none"
	IsolationContainer Isolation = "container"
	IsolationVM Isolation = "vm"
)

type Profile struct {
	ID string `json:"id"`
	ToolID string `json:"tool_id"`
	Enabled bool `json:"enabled"`
	AllowedArguments []string `json:"allowed_arguments,omitempty"`
	AllowedTargets []string `json:"allowed_targets,omitempty"`
	NetworkEnabled bool `json:"network_enabled"`
	FilesystemReadOnly bool `json:"filesystem_read_only"`
	CPUSeconds int `json:"cpu_seconds,omitempty"`
	MemoryMB int `json:"memory_mb,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`
	Isolation Isolation `json:"isolation"`
	RequiresApproval bool `json:"requires_approval"`
}
