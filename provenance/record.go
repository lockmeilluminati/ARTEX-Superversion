// Package provenance models the evidence behind tool identity, ownership and version claims.
package provenance

import "time"

type Confidence string
const (
	ConfidenceUnknown Confidence = "unknown"
	ConfidenceLow Confidence = "low"
	ConfidenceMedium Confidence = "medium"
	ConfidenceHigh Confidence = "high"
)

type EventKind string
const (
	EventDiscovered EventKind = "discovered"
	EventRelease EventKind = "release"
	EventRenamed EventKind = "renamed"
	EventTransferred EventKind = "transferred"
	EventArchived EventKind = "archived"
	EventDeleted EventKind = "deleted"
	EventRestored EventKind = "restored"
	EventSourceChanged EventKind = "source_changed"
)

type Evidence struct {
	URI string `json:"uri"`
	Digest string `json:"digest,omitempty"`
	ObservedAt time.Time `json:"observed_at"`
	Note string `json:"note,omitempty"`
}

type Record struct {
	ToolID string `json:"tool_id"`
	CanonicalURI string `json:"canonical_uri,omitempty"`
	OriginalMaintainer string `json:"original_maintainer,omitempty"`
	CurrentMaintainer string `json:"current_maintainer,omitempty"`
	FirstSeen time.Time `json:"first_seen,omitempty"`
	LatestRelease string `json:"latest_release,omitempty"`
	CommitAncestry []string `json:"commit_ancestry,omitempty"`
	Confidence Confidence `json:"confidence"`
	Evidence []Evidence `json:"evidence,omitempty"`
}

type Event struct {
	ToolID string `json:"tool_id"`
	Kind EventKind `json:"kind"`
	At time.Time `json:"at"`
	SourceURI string `json:"source_uri,omitempty"`
	Summary string `json:"summary,omitempty"`
}
