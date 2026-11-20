package discovery

import (
	"context"
	"time"
)

type RunRequest struct {
	Query Query
	RequestedAt time.Time
}

type RunResult struct {
	StartedAt time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Candidates []Candidate `json:"candidates"`
	ProviderErrors map[string]string `json:"provider_errors,omitempty"`
}

// Runner is intentionally passive until a scheduler and source providers are configured.
type Runner interface {
	Run(context.Context, RunRequest) (RunResult, error)
}
