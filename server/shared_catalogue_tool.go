package server

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/Autumn-27/artex/db"
	"github.com/Autumn-27/artex/toolregistry"
	acperm "github.com/Autumn-27/norma/permission"
	actool "github.com/Autumn-27/norma/tool"
)

type sharedCatalogueEntry struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Source               string   `json:"source"`
	Description          string   `json:"description,omitempty"`
	Enabled              bool     `json:"enabled"`
	Kind                 string   `json:"kind,omitempty"`
	Category             string   `json:"category,omitempty"`
	License              string   `json:"license,omitempty"`
	Platforms            []string `json:"platforms,omitempty"`
	Capabilities         []string `json:"capabilities,omitempty"`
	ExecutionMode        string   `json:"execution_mode"`
	VerificationStatus   string   `json:"verification_status"`
	InstallationStatus   string   `json:"installation_status"`
	RequiredCapabilities []string `json:"required_agent_capabilities,omitempty"`
	Availability         string   `json:"availability,omitempty"`
	Trust                string   `json:"trust,omitempty"`
	ApprovalRequired     bool     `json:"approval_required"`
	Lifecycle            string   `json:"lifecycle,omitempty"`
}

// sharedToolCatalogue is injected into every agent by wireAgentAugment.
// It deliberately returns metadata only: no custom exec specs, credentials,
// input schemas, handlers, or permission mutations cross this boundary.
func sharedToolCatalogue(pg *db.DB) actool.CoreTool {
	return actool.Build(actool.Spec{
		Name: "shared_tool_catalogue",
		Description: "Read-only view of the platform-wide tool catalogue, including native, custom and external profiles. Use this to discover what tools exist and understand their declared readiness. Catalogue visibility does NOT mean a tool is callable, installed, verified, approved for execution, or bound to you. Only use tools actually present in your callable tool list; never try to execute a catalogue entry by name.",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Optional case-insensitive search across tool name, description, category and capabilities"},
				"source": map[string]any{"type": "string", "enum": []string{"all", "native", "custom", "external"}, "description": "Optional source filter; defaults to all"},
				"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "description": "Maximum entries to return; defaults to 50"},
			},
		},
		ReadOnly: func(json.RawMessage) bool { return true },
		Concurrent: func(json.RawMessage) bool { return true },
		Permissions: func(context.Context, json.RawMessage, acperm.Context) acperm.Decision { return acperm.Allowed() },
		Run: func(_ context.Context, in json.RawMessage, _ *actool.ToolContext) (actool.Result, error) {
			var args struct {
				Query  string `json:"query"`
				Source string `json:"source"`
				Limit  int    `json:"limit"`
			}
			if len(in) > 0 {
				if err := json.Unmarshal(in, &args); err != nil {
					return actool.Errorf("invalid catalogue query: " + err.Error()), nil
				}
			}
			if pg == nil {
				return actool.Errorf("shared tool catalogue is unavailable: database is not initialized"), nil
			}
			if args.Limit <= 0 { args.Limit = 50 }
			if args.Limit > 100 { args.Limit = 100 }
			source := strings.ToLower(strings.TrimSpace(args.Source))
			if source == "" { source = "all" }
			if source != "all" && source != "native" && source != "custom" && source != "external" {
				return actool.Errorf("source must be all, native, custom, or external"), nil
			}
			query := strings.ToLower(strings.TrimSpace(args.Query))
			entries := make([]sharedCatalogueEntry, 0)
			runtimeTools, err := pg.ListTools()
			if err != nil { return actool.Errorf("read runtime catalogue: " + err.Error()), nil }
			for _, t := range runtimeTools {
				if t == nil { continue }
				kind := "native"
				if !t.System { kind = "custom" }
				entry := sharedCatalogueEntry{ID: "runtime." + t.Key, Name: t.Key, Source: kind, Description: t.Description, Enabled: t.Enabled, Kind: t.Kind, ExecutionMode: "unknown", VerificationStatus: "unverified", InstallationStatus: "unknown"}
				if source != "all" && source != kind { continue }
				if catalogueMatches(query, entry.Name, entry.Description, entry.Category, entry.Capabilities) { entries = append(entries, entry) }
			}
			if source == "all" || source == "external" {
				registryEntries, err := pg.ListRegistryEntries()
				if err != nil { return actool.Errorf("read external catalogue: " + err.Error()), nil }
				for _, row := range registryEntries {
					var manifest toolregistry.Manifest
					if err := json.Unmarshal(row.Manifest, &manifest); err != nil { continue }
					if manifest.ID == "" { manifest.ID = row.ID }
					if manifest.Name == "" { manifest.Name = row.Name }
					entry := sharedCatalogueEntry{
						ID: row.ID, Name: row.Name, Source: "external", Description: manifest.Description,
						Enabled: manifest.Enabled, Kind: manifest.ExecutionKind, Category: manifest.Category,
						License: manifest.License, Platforms: manifest.Platforms, Capabilities: manifest.Capabilities,
						ExecutionMode: string(manifest.ExecutionMode), VerificationStatus: string(manifest.VerificationStatus),
						InstallationStatus: string(manifest.InstallationStatus), Availability: string(manifest.Availability),
						Trust: string(manifest.Trust), ApprovalRequired: manifest.ApprovalRequired, Lifecycle: string(manifest.Lifecycle),
					}
					for _, capability := range manifest.RequiredAgentCapabilities { entry.RequiredCapabilities = append(entry.RequiredCapabilities, string(capability)) }
					if entry.ExecutionMode == "" { entry.ExecutionMode = "unknown" }
					if entry.VerificationStatus == "" { entry.VerificationStatus = "unverified" }
					if entry.InstallationStatus == "" { entry.InstallationStatus = "unknown" }
					if catalogueMatches(query, entry.Name, entry.Description, entry.Category, entry.Capabilities) { entries = append(entries, entry) }
				}
			}
			sort.Slice(entries, func(i, j int) bool {
				if entries[i].Source != entries[j].Source { return entries[i].Source < entries[j].Source }
				return entries[i].Name < entries[j].Name
			})
			total := len(entries)
			if len(entries) > args.Limit { entries = entries[:args.Limit] }
			payload, err := json.Marshal(map[string]any{
				"entries": entries, "returned": len(entries), "total_matches": total,
				"truncated": total > len(entries),
				"execution_boundary": "Catalogue visibility is descriptive only. It does not install, enable, approve, bind, or execute a tool. Use only callable tools explicitly supplied to this agent.",
			})
			if err != nil { return actool.Errorf("encode catalogue: " + err.Error()), nil }
			return actool.Text(string(payload)), nil
		},
	})
}

func catalogueMatches(query string, fields ...string) bool {
	if query == "" { return true }
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) { return true }
	}
	return false
}
