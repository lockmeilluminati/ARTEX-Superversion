// Package tooldocs models source-backed documentation drafts.
package tooldocs

import "time"

type Section struct {
	Title string `json:"title"`
	Body string `json:"body"`
	SourceURIs []string `json:"source_uris,omitempty"`
}

type Profile struct {
	ToolID string `json:"tool_id"`
	GeneratedAt time.Time `json:"generated_at"`
	Status string `json:"status"`
	Sections []Section `json:"sections"`
	Warnings []string `json:"warnings,omitempty"`
}

type Generator interface {
	Generate(toolID string) (Profile, error)
}

type DraftOnly struct{}
func (DraftOnly) Generate(toolID string) (Profile, error) {
	return Profile{ToolID: toolID, GeneratedAt: time.Now().UTC(), Status: "draft", Warnings: []string{"documentation sources are not configured"}}, nil
}
