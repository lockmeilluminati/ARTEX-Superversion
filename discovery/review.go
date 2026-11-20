package discovery

type ReviewState string
const (
	ReviewPending ReviewState = "pending"
	ReviewAccepted ReviewState = "accepted"
	ReviewRejected ReviewState = "rejected"
	ReviewQuarantined ReviewState = "quarantined"
)

type Review struct {
	CandidateID string `json:"candidate_id"`
	State ReviewState `json:"state"`
	Reviewer string `json:"reviewer,omitempty"`
	Reason string `json:"reason,omitempty"`
	ReviewedAt string `json:"reviewed_at,omitempty"`
}

// Review is a human trust decision. Candidate discovery alone must never change it.
