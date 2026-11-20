package agent

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Autumn-27/artex/normalizer"
)

// observationContext identifies one ingestion of a retained raw artifact. The
// run ID is explicitly an ARTEX ingestion-run ID, not a claim about the
// scanner's own process/run identifier (which these file formats do not carry).
type observationContext struct {
	TaskID     int64
	SourceTool string
	Artifact   string
	RunID      string
	ObservedAt time.Time
}

func newObservationContext(taskID int64, sourceTool, path string, now time.Time) (observationContext, error) {
	if taskID <= 0 {
		return observationContext{}, fmt.Errorf("task ID unavailable; normalized observations require task provenance")
	}
	if strings.TrimSpace(sourceTool) == "" || strings.TrimSpace(path) == "" {
		return observationContext{}, fmt.Errorf("source tool and artifact path are required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	// Nanosecond timestamp makes each ingestion attempt distinct. Exact replay
	// deduplication is still handled by the source fact/asset path; scanner run
	// identity must be supplied by a future execution-aware runner integration.
	runID := fmt.Sprintf("ingest-task-%d-%d", taskID, now.UnixNano())
	return observationContext{
		TaskID: taskID, SourceTool: sourceTool,
		Artifact: "artifact://" + filepath.Base(path),
		RunID: runID, ObservedAt: now,
	}, nil
}

func observationFromAsset(ctx observationContext, item assetInputItem) (normalizer.Observation, bool) {
	o := normalizer.Observation{
		SourceToolID: ctx.SourceTool, SourceRunID: ctx.RunID, TaskID: ctx.TaskID,
		ObservedAt: ctx.ObservedAt, Evidence: []normalizer.EvidenceReference{{
			Kind: "raw_output", URI: ctx.Artifact, ObservedAt: ctx.ObservedAt,
		}},
		Scope: normalizer.ScopeUnknown, Confidence: normalizer.ConfidenceHigh,
		Verification: normalizer.VerificationUnverified,
		Collection: normalizer.CollectionTestedPositive,
	}
	switch {
	case item.URL != "":
		o.Type, o.Target = normalizer.ObservationURL, item.URL
		if len(item.Technologies) > 0 {
			o.Attributes = map[string]string{"technologies": strings.Join(item.Technologies, ",")}
		}
		if item.StatusCode != nil {
			o.Attributes = ensureObservationAttributes(o.Attributes)
			o.Attributes["http_status"] = fmt.Sprint(*item.StatusCode)
		}
	case item.ServiceIP != "" && item.Port > 0 && item.Port <= 65535:
		o.Type, o.Target = normalizer.ObservationPort, item.ServiceIP
		p := uint16(item.Port)
		o.Port = &p
		o.Protocol = item.Proto
	case item.IP != "" && item.Port > 0 && item.Port <= 65535:
		o.Type, o.Target = normalizer.ObservationPort, item.IP
		p := uint16(item.Port)
		o.Port, o.Protocol = &p, item.Proto
	case item.IP != "":
		o.Type, o.Target = normalizer.ObservationIP, item.IP
	case item.Domain != "":
		o.Type, o.Target = normalizer.ObservationDomain, item.Domain
	default:
		return normalizer.Observation{}, false
	}
	if item.ServiceName != "" {
		o.Attributes = ensureObservationAttributes(o.Attributes)
		o.Attributes["service_name"] = item.ServiceName
	}
	if err := o.AssignIDs(); err != nil || o.Validate() != nil {
		return normalizer.Observation{}, false
	}
	return o, true
}

func ensureObservationAttributes(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	return in
}

func scanObservations(ctx observationContext, assets []assetInputItem) []normalizer.Observation {
	out := make([]normalizer.Observation, 0, len(assets))
	for _, item := range assets {
		if o, ok := observationFromAsset(ctx, item); ok {
			out = append(out, o)
		}
	}
	return out
}

func bbotObservations(ctx observationContext, events []bbotEvent, inScope bool) []normalizer.Observation {
	out := make([]normalizer.Observation, 0, len(events))
	for _, ev := range events {
		target := ev.Host
		typ := normalizer.ObservationHost
		switch {
		case ev.Type == "DNS_NAME" && target != "":
			typ = normalizer.ObservationDomain
		case ev.Type == "IP_ADDRESS" && target != "":
			typ = normalizer.ObservationIP
		case ev.Type == "URL" && ev.URL != "":
			typ, target = normalizer.ObservationURL, ev.URL
		case ev.Type == "OPEN_TCP_PORT" && target != "" && ev.Port > 0:
			typ = normalizer.ObservationPort
		default:
			continue
		}
		o := normalizer.Observation{
			Type: typ, Target: target, SourceToolID: ctx.SourceTool,
			SourceRunID: ctx.RunID, TaskID: ctx.TaskID, ObservedAt: ctx.ObservedAt,
			Evidence: []normalizer.EvidenceReference{{Kind: "raw_output", URI: ctx.Artifact, ObservedAt: ctx.ObservedAt}},
			Scope: normalizer.ScopeOutOfScope, Confidence: normalizer.ConfidenceMedium,
			Verification: normalizer.VerificationUnverified, Collection: normalizer.CollectionTestedPositive,
			Attributes: map[string]string{"event_type": ev.Type, "module": ev.Module},
		}
		if inScope {
			o.Scope = normalizer.ScopeInScope
		}
		if typ == normalizer.ObservationPort {
			p := uint16(ev.Port)
			o.Port, o.Protocol = &p, "tcp"
		}
		if err := o.AssignIDs(); err != nil || o.Validate() != nil {
			continue
		}
		out = append(out, o)
	}
	return out
}
