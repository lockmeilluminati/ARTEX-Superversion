package toolregistry

import (
	"encoding/json"
	"strings"

	"github.com/Autumn-27/artex/db"
)

// ImportDBTools converts the existing ARTEX tools table into registry metadata.
// It is deliberately a projection: it neither changes rows nor creates handlers.
// Native handler identity stays in Go; custom execution specs remain owned by db/server.
func ImportDBTools(rows []*db.Tool) []Manifest {
	out := make([]Manifest, 0, len(rows))
	for _, row := range rows {
		if row == nil || strings.TrimSpace(row.Key) == "" {
			continue
		}
		out = append(out, FromDBTool(row))
	}
	return out
}

// FromDBTool maps one legacy ARTEX tool row to a stable namespaced registry ID.
func FromDBTool(row *db.Tool) Manifest {
	if row == nil {
		return Manifest{}
	}
	sourceKind, executionKind, trust := "artex_custom", row.Kind, TrustOperatorConfigured
	id := "artex.custom." + row.Key
	if row.System {
		sourceKind, executionKind, trust = "artex_native", "native_handler", TrustCodeDefined
		id = "artex.native." + row.Key
	}
	if executionKind == "" {
		executionKind = "native_handler"
	}
	if executionKind == "shell" {
		// A shell-kind custom tool is a model-facing environment declaration,
		// not a standalone executor registered by this manifest.
		executionKind = "shell_declaration"
	}
	schema := json.RawMessage(row.Schema)
	if len(schema) == 0 {
		schema = json.RawMessage("{}")
	}
	return Manifest{
		ID: id, LegacyKey: row.Key, Name: row.Key, Description: row.Description,
		Sources: []Source{{Kind: sourceKind}},
		AgentBindings: append([]string(nil), row.Agents...),
		ExecutionKind: executionKind, InputSchema: schema, Enabled: row.Enabled,
		Availability: AvailabilityRegistered, Trust: trust,
		ApprovalRequired: true, Lifecycle: LifecycleAvailable, Verified: row.System,
	}
}
