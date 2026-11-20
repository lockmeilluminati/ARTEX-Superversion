package agent

import (
	"testing"
	"time"

	"github.com/Autumn-27/artex/normalizer"
)

func testObservationContext() observationContext {
	at := time.Date(2026, 11, 20, 14, 0, 0, 0, time.UTC)
	return observationContext{TaskID: 42, SourceTool: "external.httpx", Artifact: "artifact://scan.jsonl", RunID: "ingest-task-42-1", ObservedAt: at}
}

func TestObservationFromHttpxAssetPreservesProvenance(t *testing.T) {
	status := 200
	o, ok := observationFromAsset(testObservationContext(), assetInputItem{
		Type: "service", URL: "https://app.example.test", StatusCode: &status,
		Technologies: []string{"nginx", "react"},
	})
	if !ok {
		t.Fatal("expected URL observation")
	}
	if err := o.Validate(); err != nil {
		t.Fatalf("invalid normalized observation: %v", err)
	}
	if o.SourceToolID != "external.httpx" || o.SourceRunID != "ingest-task-42-1" || o.TaskID != 42 {
		t.Fatalf("provenance lost: %+v", o)
	}
	if o.Evidence[0].URI != "artifact://scan.jsonl" || o.Attributes["http_status"] != "200" {
		t.Fatalf("evidence or typed attributes lost: %+v", o)
	}
	if o.Verification != normalizer.VerificationUnverified {
		t.Fatalf("scanner observation must remain unverified, got %q", o.Verification)
	}
}

func TestObservationFromNmapServiceCreatesPortObservation(t *testing.T) {
	o, ok := observationFromAsset(testObservationContext(), assetInputItem{
		Type: "service", IP: "192.0.2.10", Port: 443, Proto: "tcp", ServiceName: "https",
	})
	if !ok {
		t.Fatal("expected port observation")
	}
	if o.Type != normalizer.ObservationPort || o.Port == nil || *o.Port != 443 || o.Protocol != "tcp" {
		t.Fatalf("unexpected port observation: %+v", o)
	}
}

func TestBBotObservationsPreserveScopeDistinction(t *testing.T) {
	ctx := testObservationContext()
	ctx.SourceTool = "external.bbot"
	events := []bbotEvent{
		{Type: "DNS_NAME", Host: "in.example.test"},
		{Type: "DNS_NAME", Host: "third-party.example.test"},
	}
	in := bbotObservations(ctx, events[:1], true)
	out := bbotObservations(ctx, events[1:], false)
	if len(in) != 1 || in[0].Scope != normalizer.ScopeInScope {
		t.Fatalf("in-scope event not represented correctly: %+v", in)
	}
	if len(out) != 1 || out[0].Scope != normalizer.ScopeOutOfScope {
		t.Fatalf("third-party event not represented correctly: %+v", out)
	}
	if in[0].SourceToolID != "external.bbot" || out[0].SourceToolID != "external.bbot" {
		t.Fatal("BBot source attribution missing")
	}
}

func TestNewObservationContextRequiresTaskProvenance(t *testing.T) {
	if _, err := newObservationContext(0, "external.nmap", "out.xml", time.Now()); err == nil {
		t.Fatal("expected missing task provenance to fail closed")
	}
}
