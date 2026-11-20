package veq

import "context"

type TaskRequirements struct {
	Capabilities []string
	Platforms []string
	AllowedPermissions []string
}

type Recommendation struct {
	ToolID string `json:"tool_id"`
	Score float64 `json:"score"`
	Reasons []string `json:"reasons"`
	Warnings []string `json:"warnings,omitempty"`
	RequiresApproval bool `json:"requires_approval"`
}

type Recommender interface {
	Recommend(context.Context, TaskRequirements) ([]Recommendation, error)
}

// NoopRecommender intentionally returns no executable recommendations until
// verified registry and policy data are connected.
type NoopRecommender struct{}
func (NoopRecommender) Recommend(context.Context, TaskRequirements) ([]Recommendation, error) {
	return []Recommendation{}, nil
}
