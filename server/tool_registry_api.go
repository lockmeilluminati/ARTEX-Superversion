package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

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
	ReviewStatus string `json:"review_status"`
	ReviewNote string `json:"review_note,omitempty"`
	ReviewedBy string `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
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
	usageCounts, usageErr := pg.ToolUsageCounts()
	if usageErr != nil {
		log.Printf("[tool-registry] usage counts unavailable: %v", usageErr)
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
			Kind: tool.Kind, Agents: tool.Agents, Calls: usageCounts[tool.Key],
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
			Manifest: &manifest, ReviewStatus: row.ReviewStatus,
			ReviewNote: row.ReviewNote, ReviewedBy: row.ReviewedBy, ReviewedAt: row.ReviewedAt,
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


type registryReviewRequest struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

func validateRegistryReviewRequest(req registryReviewRequest) error {
	req.Status = strings.TrimSpace(req.Status)
	req.Note = strings.TrimSpace(req.Note)
	if req.Status != "pending" && req.Status != "approved" && req.Status != "rejected" {
		return errors.New("status must be pending, approved, or rejected")
	}
	if req.Status == "rejected" && req.Note == "" {
		return errors.New("a reason is required when rejecting a profile")
	}
	if len(req.Note) > 2000 {
		return errors.New("review note must be 2000 characters or fewer")
	}
	return nil
}

// pgReviewToolRegistry records a decision by the authenticated ARTEX operator.
// Approval is a metadata review only; it never enables, installs, or grants execution.
func (s *Server) pgReviewToolRegistry(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if !strings.HasPrefix(id, "external.") {
		writeErr(w, http.StatusBadRequest, "only external registry profiles can be reviewed")
		return
	}
	reviewer := s.approver(r)
	if reviewer == "" {
		writeErr(w, http.StatusUnauthorized, "valid operator session required")
		return
	}
	var req registryReviewRequest
	if err := decode(w, r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Status, req.Note = strings.TrimSpace(req.Status), strings.TrimSpace(req.Note)
	if err := validateRegistryReviewRequest(req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := pg.ReviewRegistryEntry(id, req.Status, req.Note, reviewer); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, err.Error())
		} else if errors.Is(err, db.ErrInvalidRegistryReview) {
			writeErr(w, http.StatusBadRequest, err.Error())
		} else {
			writeErr(w, http.StatusInternalServerError, "could not save registry review")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "review_status": req.Status})
}

func (s *Server) pgListToolRegistryReviews(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil { return }
	id := strings.TrimSpace(r.PathValue("id"))
	if !strings.HasPrefix(id, "external.") {
		writeErr(w, http.StatusBadRequest, "only external registry profiles have review history")
		return
	}
	reviews, err := pg.ListRegistryReviews(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not load registry review history")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reviews": reviews})
}
