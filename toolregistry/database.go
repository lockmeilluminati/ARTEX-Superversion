package toolregistry

import "github.com/Autumn-27/artex/db"

// LoadFromDB builds an in-memory registry view from the existing ARTEX tool table.
// It is read-only and intentionally does not persist registry metadata or alter tool rows.
func LoadFromDB(pg *db.DB) (*Registry, error) {
	if pg == nil {
		return nil, ErrNilDB
	}
	rows, err := pg.ListTools()
	if err != nil {
		return nil, err
	}
	reg := New()
	for _, manifest := range ImportDBTools(rows) {
		if err := reg.Register(manifest); err != nil {
			return nil, err
		}
	}
	return reg, nil
}
