package db

import (
	"errors"
	"strings"
	"time"
)

// ToolCollection is an operator-defined grouping of catalogue IDs. It has no
// effect on tool installation, enablement, permissions, or execution.
type ToolCollection struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ToolIDs     []string  `json:"tool_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var ErrInvalidToolCollection = errors.New("collection name is required")

func (d *DB) CreateToolCollection(name, description string, toolIDs []string) (int64, error) {
	if d == nil || d.DB == nil {
		return 0, errors.New("database is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, ErrInvalidToolCollection
	}
	var id int64
	err := d.QueryRow(
		`INSERT INTO tool_collections (name, description) VALUES ($1, $2) RETURNING id`,
		name, strings.TrimSpace(description),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	if err := d.ReplaceToolCollectionItems(id, toolIDs); err != nil {
		_, _ = d.Exec(`DELETE FROM tool_collections WHERE id = $1`, id)
		return 0, err
	}
	return id, nil
}

func (d *DB) ListToolCollections() ([]ToolCollection, error) {
	if d == nil || d.DB == nil {
		return nil, errors.New("database is required")
	}
	rows, err := d.Query(`
SELECT c.id, c.name, c.description, c.created_at, c.updated_at,
       COALESCE(array_agg(i.tool_id ORDER BY i.position, i.tool_id)
                FILTER (WHERE i.tool_id IS NOT NULL), ARRAY[]::text[])
FROM tool_collections c
LEFT JOIN tool_collection_items i ON i.collection_id = c.id
GROUP BY c.id
ORDER BY lower(c.name), c.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ToolCollection
	for rows.Next() {
		var item ToolCollection
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt, &item.ToolIDs); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (d *DB) UpdateToolCollection(id int64, name, description string, toolIDs []string) error {
	if d == nil || d.DB == nil {
		return errors.New("database is required")
	}
	name = strings.TrimSpace(name)
	if id <= 0 || name == "" {
		return ErrInvalidToolCollection
	}
	result, err := d.Exec(`UPDATE tool_collections SET name=$2, description=$3, updated_at=now() WHERE id=$1`,
		id, name, strings.TrimSpace(description))
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("tool collection not found")
	}
	return d.ReplaceToolCollectionItems(id, toolIDs)
}

func (d *DB) ReplaceToolCollectionItems(id int64, toolIDs []string) error {
	if d == nil || d.DB == nil {
		return errors.New("database is required")
	}
	if id <= 0 {
		return ErrInvalidToolCollection
	}
	if _, err := d.Exec(`DELETE FROM tool_collection_items WHERE collection_id=$1`, id); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(toolIDs))
	for position, toolID := range toolIDs {
		toolID = strings.TrimSpace(toolID)
		if toolID == "" {
			continue
		}
		if _, exists := seen[toolID]; exists {
			continue
		}
		seen[toolID] = struct{}{}
		if _, err := d.Exec(`INSERT INTO tool_collection_items (collection_id, tool_id, position) VALUES ($1,$2,$3)`,
			id, toolID, position); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) DeleteToolCollection(id int64) error {
	if d == nil || d.DB == nil {
		return errors.New("database is required")
	}
	if id <= 0 {
		return ErrInvalidToolCollection
	}
	result, err := d.Exec(`DELETE FROM tool_collections WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("tool collection not found")
	}
	return nil
}
