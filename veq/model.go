// Package veq is the planned intelligence layer over ARTEX's existing asset,
// exploration and evidence graphs. It does not create a parallel execution engine.
package veq

import "time"

type NodeKind string
const (
	NodeTool NodeKind = "tool"
	NodeVersion NodeKind = "version"
	NodeSource NodeKind = "source"
	NodeTechnique NodeKind = "technique"
	NodeAsset NodeKind = "asset"
	NodeFinding NodeKind = "finding"
	NodeEvidence NodeKind = "evidence"
)

type Node struct {
	ID string `json:"id"`
	Kind NodeKind `json:"kind"`
	Label string `json:"label"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Edge struct {
	From string `json:"from"`
	To string `json:"to"`
	Relation string `json:"relation"`
	Confidence float64 `json:"confidence,omitempty"`
}

type Snapshot struct {
	GeneratedAt time.Time `json:"generated_at"`
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}
