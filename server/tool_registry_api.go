package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Autumn-27/artex/db"
	"github.com/Autumn-27/artex/toolregistry"
)

// registryCatalogRecord is a read-only marketplace projection. Registry metadata
// describes a tool; it never creates or grants a runtime execution handler.
type registryCatalogRecord struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	Manifest toolregistry.Manifest `json:"manifest"`
}

// pgListToolRegistry exposes curated external-tool metadata to the marketplace.
// Native and custom runtime tools remain available through GET /api/tools.
func (s *Server) pgListToolRegistry(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	rows, err := pg.ListRegistryEntries()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	entries := make([]registryCatalogRecord, 0, len(rows))
	for _, row := range rows {
		var manifest toolregistry.Manifest
		if err := json.Unmarshal(row.Manifest, &manifest); err != nil {
			// A malformed catalogue record should be visible in server logs, but
			// must not prevent other valid tools from appearing in the marketplace.
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
			ID: row.ID, Name: row.Name, Manifest: manifest,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
}

// Keep the DB type referenced here so this endpoint's response contract remains
// explicitly tied to the persisted registry row model.
var _ = db.RegistryEntry{}
