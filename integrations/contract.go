package integrations

import (
	"context"
	"time"
)

type Operation struct {
	ID string `json:"id"`
	IntegrationID string `json:"integration_id"`
	Capability string `json:"capability"`
	Target string `json:"target,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
	RequestedBy string `json:"requested_by,omitempty"`
	ApprovalID string `json:"approval_id,omitempty"`
}

type OperationResult struct {
	OperationID string `json:"operation_id"`
	IntegrationID string `json:"integration_id"`
	Status string `json:"status"`
	StartedAt time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
	Summary string `json:"summary,omitempty"`
	EvidenceRefs []string `json:"evidence_refs,omitempty"`
}

// Executor is a future adapter contract. Implementations must enforce ARTEX scope
// and approval before any operation with side effects.
type Executor interface {
	Execute(context.Context, Operation) (OperationResult, error)
}
