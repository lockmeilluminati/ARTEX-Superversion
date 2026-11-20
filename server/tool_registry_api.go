package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"

	"github.com/Autumn-27/artex/db"
	"github.com/Autumn-27/artex/toolregistry"
)

// registryCatalogRecord is the safe, read-only marketplace projection. It
// deliberately excludes custom-tool execution specs and never grants a handler.
type registryCatalogRecord struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Source      string                 `json:"source"`
	Description string                 `json:"description,omitempty"`
	Enabled     bool                   `json:"enabled"`
	Kind        string                 `json:"kind,omitempty"`
	Agents      []string               `json:"agents,omitempty"`
	Calls       int                    `json:"calls,omitempty"`
	Manifest    *toolregistry.Manifest `json:"manifest,omitempty"`
}

// pgListToolRegistry returns one catalogue containing legacy native/custom
// runtime tools and separately registered external tools. The projection is
// descriptive only: catalogue membership is not installation or execution.
func (s *Server) pgListToolRegistry(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}

	tools, err := pg.ListTools()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	entries := make([]registryCatalogRecord, 0, len(tools))
	for _, tool := range tools {
		if tool == nil {
			continue
		}
		source := "native"
		if !tool.System {
			source = "custom"
		}
		entries = append(entries, registryCatalogRecord{
			ID: "runtime." + tool.Key, Name: tool.Key, Source: source,
			Description: tool.Description, Enabled: tool.Enabled,
			Kind: tool.Kind, Agents: tool.Agents, Calls: tool.Calls,
		})
	}

	rows, err := pg.ListRegistryEntries()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, row := range rows {
		var manifest toolregistry.Manifest
		if err := json.Unmarshal(row.Manifest, &manifest); err != nil {
			log.Printf("[tool-registry] skipping malformed manifest %q: %v", row.ID, err)
			continue
		}
		if manifest.ID == "" {
			manifest.ID = row.ID
		}
		if manifest.Name == "" {
			manifest.Name = row.Name
		}
		entries = append(entries, registryCatalogRecord{
			ID: row.ID, Name: row.Name, Source: "external",
			Description: manifest.Description, Enabled: manifest.Enabled,
			Kind: manifest.ExecutionKind, Agents: manifest.AgentBindings,
			Manifest: &manifest,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Source != entries[j].Source {
			return entries[i].Source < entries[j].Source
		}
		return entries[i].Name < entries[j].Name
	})
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
}

// Keep the DB row contract explicit in this package's compile-time surface.
var _ db.RegistryEntry
