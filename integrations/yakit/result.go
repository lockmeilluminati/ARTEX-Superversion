package yakit

import "time"

// ResultEnvelope is an ARTEX-side normalized boundary, not a claim about Yakit's
// native wire format. A version-specific translator must populate it after the
// actual Yakit interface and schema have been verified.
type ResultEnvelope struct {
	IntegrationVersion string `json:"integration_version,omitempty"`
	OperationID string `json:"operation_id"`
	ObservedAt time.Time `json:"observed_at"`
	Findings []Finding `json:"findings,omitempty"`
	Traffic []TrafficReference `json:"traffic,omitempty"`
}

type Finding struct {
	Title string `json:"title"`
	Severity string `json:"severity,omitempty"`
	Target string `json:"target,omitempty"`
	Description string `json:"description,omitempty"`
	NativeReference string `json:"native_reference,omitempty"`
}

type TrafficReference struct {
	RequestReference string `json:"request_reference,omitempty"`
	ResponseReference string `json:"response_reference,omitempty"`
	NativeReference string `json:"native_reference,omitempty"`
}
