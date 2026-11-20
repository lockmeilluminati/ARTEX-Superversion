package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// RegistryEntry stores descriptive metadata for a tool independently from the
// legacy tools table. It does not define or grant an execution handler.
type RegistryEntry struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Manifest     json.RawMessage `json:"manifest"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	ReviewStatus string          `json:"review_status"`
	ReviewNote   string          `json:"review_note"`
	ReviewedBy   string          `json:"reviewed_by"`
	ReviewedAt   *time.Time      `json:"reviewed_at,omitempty"`
}

var ErrInvalidRegistryEntry = errors.New("registry entry requires an ID, name, and valid JSON manifest")

// UpsertRegistryEntry inserts or refreshes descriptive metadata by stable ID.
// Runtime tool rows and execution permissions are intentionally untouched.
func (d *DB) UpsertRegistryEntry(entry RegistryEntry) error {
	if d == nil || d.DB == nil {
		return errors.New("database is required")
	}
	if strings.TrimSpace(entry.ID) == "" || strings.TrimSpace(entry.Name) == "" || len(entry.Manifest) == 0 || !json.Valid(entry.Manifest) {
		return ErrInvalidRegistryEntry
	}
	_, err := d.Exec(`
INSERT INTO tool_registry (id, name, manifest)
VALUES ($1, $2, $3::jsonb)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name, manifest = EXCLUDED.manifest, updated_at = now()`,
		entry.ID, entry.Name, string(entry.Manifest))
	return err
}

// ListRegistryEntries returns descriptive tool metadata in stable ID order.
func (d *DB) ListRegistryEntries() ([]RegistryEntry, error) {
	if d == nil || d.DB == nil {
		return nil, errors.New("database is required")
	}
	rows, err := d.Query(`SELECT id, name, manifest, created_at, updated_at, review_status, review_note, reviewed_by, reviewed_at FROM tool_registry ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []RegistryEntry
	for rows.Next() {
		var entry RegistryEntry
		var reviewedAt sql.NullTime
		if err := rows.Scan(&entry.ID, &entry.Name, &entry.Manifest, &entry.CreatedAt, &entry.UpdatedAt, &entry.ReviewStatus, &entry.ReviewNote, &entry.ReviewedBy, &reviewedAt); err != nil {
			return nil, err
		}
		if reviewedAt.Valid {
			entry.ReviewedAt = &reviewedAt.Time
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
