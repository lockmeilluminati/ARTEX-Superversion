package veq

import "context"

type GraphReader interface {
	Snapshot(context.Context) (Snapshot, error)
}

type Index interface {
	Upsert(context.Context, Node) error
	Relate(context.Context, Edge) error
	Search(context.Context, string, int) ([]Node, error)
}

// Indexing is a projection over ARTEX's existing records. The implementation
// must retain source record IDs and must not become a second source of truth.
