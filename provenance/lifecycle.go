package provenance

import "time"

// LifecycleObservation records what a source reported at a point in time.
// It is an observation, not proof that a repository was permanently deleted.
type LifecycleObservation struct {
	ToolID string `json:"tool_id"`
	RepositoryURI string `json:"repository_uri"`
	Kind EventKind `json:"kind"`
	ObservedAt time.Time `json:"observed_at"`
	SourceURI string `json:"source_uri"`
	Confidence Confidence `json:"confidence"`
	Details string `json:"details,omitempty"`
}
