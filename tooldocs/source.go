package tooldocs

import "time"

type SourceDocument struct {
	URI string `json:"uri"`
	Title string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Digest string `json:"digest,omitempty"`
	RetrievedAt time.Time `json:"retrieved_at"`
	Publisher string `json:"publisher,omitempty"`
	Official bool `json:"official"`
}

type SourceCollector interface {
	Collect(toolID string) ([]SourceDocument, error)
}
