package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Autumn-27/artex/db"
)

type toolCollectionRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ToolIDs     []string `json:"tool_ids"`
}

// validateToolCollectionIDs accepts only stable IDs exposed by the read-only
// marketplace projection. Collection membership never grants execution access.
func validateToolCollectionIDs(available map[string]struct{}, ids []string) ([]string, error) {
	validated := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, fmt.Errorf("tool_ids cannot contain an empty ID")
		}
		if _, ok := available[id]; !ok {
			return nil, fmt.Errorf("unknown tool ID %q", id)
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		validated = append(validated, id)
	}
	return validated, nil
}

func (s *Server) availableToolCollectionIDs(pg *db.DB) (map[string]struct{}, error) {
	available := make(map[string]struct{})
	tools, err := pg.ListTools()
	if err != nil {
		return nil, err
	}
	for _, tool := range tools {
		if tool != nil {
			available["runtime."+tool.Key] = struct{}{}
		}
	}
	entries, err := pg.ListRegistryEntries()
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		available[entry.ID] = struct{}{}
	}
	return available, nil
}

func (s *Server) validateToolCollectionRequest(pg *db.DB, req *toolCollectionRequest) error {
	available, err := s.availableToolCollectionIDs(pg)
	if err != nil {
		return err
	}
	req.ToolIDs, err = validateToolCollectionIDs(available, req.ToolIDs)
	return err
}

func (s *Server) pgListToolCollections(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	collections, err := pg.ListToolCollections()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if collections == nil {
		collections = []db.ToolCollection{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"collections": collections})
}

func (s *Server) pgCreateToolCollection(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	var req toolCollectionRequest
	if err := decode(w, r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeErr(w, http.StatusBadRequest, db.ErrInvalidToolCollection.Error())
		return
	}
	if err := s.validateToolCollectionRequest(pg, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := pg.CreateToolCollection(req.Name, req.Description, req.ToolIDs)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (s *Server) pgUpdateToolCollection(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid collection ID")
		return
	}
	var req toolCollectionRequest
	if err := decode(w, r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.validateToolCollectionRequest(pg, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := pg.UpdateToolCollection(id, req.Name, req.Description, req.ToolIDs); err != nil {
		status := http.StatusInternalServerError
		if err == db.ErrInvalidToolCollection {
			status = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		writeErr(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) pgDeleteToolCollection(w http.ResponseWriter, r *http.Request) {
	pg := s.pg(w)
	if pg == nil {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid collection ID")
		return
	}
	if err := pg.DeleteToolCollection(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": id})
}
