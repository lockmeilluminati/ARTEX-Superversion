package executionprofile

import "errors"

var (
	ErrProfileIDRequired = errors.New("profile ID and tool ID are required")
	ErrExecutionDisabled = errors.New("execution profile must remain disabled until policy integration is complete")
	ErrApprovalRequired = errors.New("external tool execution must require approval")
)

func (p Profile) ValidateScaffold() error {
	if p.ID == "" || p.ToolID == "" { return ErrProfileIDRequired }
	if p.Enabled { return ErrExecutionDisabled }
	if !p.RequiresApproval { return ErrApprovalRequired }
	return nil
}
