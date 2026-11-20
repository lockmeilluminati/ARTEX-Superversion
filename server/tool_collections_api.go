package server

import (
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
