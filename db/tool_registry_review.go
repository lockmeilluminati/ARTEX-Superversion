package db

import (
	"errors"
	"strings"
	"time"
)

type RegistryReview struct {
	ID         int64     `json:"id"`
	RegistryID string    `json:"registry_id"`
	Status     string    `json:"status"`
	Note       string    `json:"note"`
	ReviewedBy string    `json:"reviewed_by"`
	ReviewedAt time.Time `json:"reviewed_at"`
}

var ErrInvalidRegistryReview = errors.New("review requires a valid registry ID, status, and reviewer")

// ReviewRegistryEntry records a metadata-only decision and appends an immutable
// history row in the same transaction. It deliberately never edits manifest JSON.
func (d *DB) ReviewRegistryEntry(registryID, status, note, reviewer string) error {
	if d == nil || d.DB == nil {
		return errors.New("database is required")
	}
	registryID = strings.TrimSpace(registryID)
	reviewer = strings.TrimSpace(reviewer)
	note = strings.TrimSpace(note)
	if !strings.HasPrefix(registryID, "external.") || reviewer == "" ||
		(status != "pending" && status != "approved" && status != "rejected") ||
		(status == "rejected" && note == "") {
		return ErrInvalidRegistryReview
	}

	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
UPDATE tool_registry
SET review_status = $2, review_note = $3, reviewed_by = $4, reviewed_at = now(), updated_at = now()
WHERE id = $1`, registryID, status, note, reviewer)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("registry entry not found")
	}
	if _, err := tx.Exec(`
INSERT INTO tool_registry_reviews (registry_id, status, note, reviewed_by)
VALUES ($1, $2, $3, $4)`, registryID, status, note, reviewer); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *DB) ListRegistryReviews(registryID string) ([]RegistryReview, error) {
	if d == nil || d.DB == nil {
		return nil, errors.New("database is required")
	}
	rows, err := d.Query(`
SELECT id, registry_id, status, note, reviewed_by, reviewed_at
FROM tool_registry_reviews
WHERE registry_id = $1
ORDER BY reviewed_at DESC, id DESC
LIMIT 100`, strings.TrimSpace(registryID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reviews := make([]RegistryReview, 0)
	for rows.Next() {
		var review RegistryReview
		if err := rows.Scan(&review.ID, &review.RegistryID, &review.Status, &review.Note, &review.ReviewedBy, &review.ReviewedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}
