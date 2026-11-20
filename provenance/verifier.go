package provenance

import "context"

type Claim struct {
	ToolID string
	RepositoryURI string
	Release string
	Maintainer string
}

type Verification struct {
	Confidence Confidence
	Signals []string
	Warnings []string
	Evidence []Evidence
}

// Verifier implementations must cite the source for each signal.
// A nil verifier is not a trust decision; candidates remain unverified.
type Verifier interface {
	Verify(context.Context, Claim) (Verification, error)
}

type UnverifiedVerifier struct{}
func (UnverifiedVerifier) Verify(context.Context, Claim) (Verification, error) {
	return Verification{Confidence: ConfidenceUnknown, Warnings: []string{"no provenance provider configured"}}, nil
}
