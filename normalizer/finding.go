// Package normalizer translates adapter output into ARTEX-neutral records.
// Persistence into ARTEX findings/evidence is a separate integration step.
package normalizer

import "time"

type ExternalFinding struct {
	SourceToolID string
	SourceToolVersion string
	Title string
	Description string
	Severity string
	Target string
	ObservedAt time.Time
	RawReference string
}

type EvidenceReference struct {
	Kind string `json:"kind"`
	URI string `json:"uri,omitempty"`
	Digest string `json:"digest,omitempty"`
	ObservedAt time.Time `json:"observed_at"`
}

type Finding struct {
	Title string `json:"title"`
	Description string `json:"description,omitempty"`
	Severity string `json:"severity,omitempty"`
	Asset string `json:"asset,omitempty"`
	SourceToolID string `json:"source_tool_id"`
	SourceToolVersion string `json:"source_tool_version,omitempty"`
	ObservedAt time.Time `json:"observed_at"`
	Evidence []EvidenceReference `json:"evidence,omitempty"`
}

func Normalize(in ExternalFinding) Finding {
	return Finding{
		Title: in.Title, Description: in.Description, Severity: in.Severity,
		Asset: in.Target, SourceToolID: in.SourceToolID,
		SourceToolVersion: in.SourceToolVersion, ObservedAt: in.ObservedAt,
		Evidence: []EvidenceReference{{Kind:"external_result", URI:in.RawReference, ObservedAt:in.ObservedAt}},
	}
}
