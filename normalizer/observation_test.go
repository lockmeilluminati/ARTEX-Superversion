package normalizer

import (
	"strings"
	"testing"
	"time"
)

func validObservation() Observation {
	return Observation{
		Type: ObservationService, Target: "https://example.test", Name: "http",
		SourceToolID: "external.httpx", SourceRunID: "run-123", TaskID: 42,
		ObservedAt: time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC),
		Evidence: []EvidenceReference{{Kind: "raw_output", URI: "evidence://run-123/httpx.jsonl", Digest: "sha256:abc", ObservedAt: time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)}},
		Scope: ScopeInScope, Confidence: ConfidenceHigh,
		Verification: VerificationUnverified, Collection: CollectionTestedPositive,
	}
}

func TestObservationValidateAcceptsExplicitUnknownStates(t *testing.T) {
	o := validObservation()
	o.Scope, o.Confidence = ScopeUnknown, ConfidenceUnknown
	o.Verification, o.Collection = VerificationNotAssessed, CollectionUnknown
	if err := o.Validate(); err != nil { t.Fatalf("explicit unknown state rejected: %v", err) }
}

func TestObservationValidateRequiresProvenanceAndEvidenceShape(t *testing.T) {
	o := validObservation()
	o.SourceRunID = ""
	if err := o.Validate(); err == nil || !strings.Contains(err.Error(), "source_run_id") { t.Fatalf("expected run provenance error, got %v", err) }
	o = validObservation()
	o.Evidence = []EvidenceReference{{Kind: "raw_output"}}
	if err := o.Validate(); err == nil || !strings.Contains(err.Error(), "evidence[0]") { t.Fatalf("expected evidence reference error, got %v", err) }
}

func TestObservationValidateDoesNotTreatNotTestedAsNegative(t *testing.T) {
	o := validObservation()
	o.Collection = CollectionNotTested
	if err := o.Validate(); err != nil { t.Fatal(err) }
	if o.Collection == CollectionTestedNegative { t.Fatal("not_tested was collapsed into tested_negative") }
}

func TestNormalizeIdentityCanonicalizesDomainsAndIPs(t *testing.T) {
	a, err := NormalizeIdentity(ObservationDomain, "Example.TEST.", nil, "", "")
	if err != nil { t.Fatal(err) }
	b, err := NormalizeIdentity(ObservationDomain, "example.test", nil, "", "")
	if err != nil { t.Fatal(err) }
	if a != b { t.Fatalf("domain identity differs: %s != %s", a, b) }
	a, err = NormalizeIdentity(ObservationIP, "2001:0db8::1", nil, "", "")
	if err != nil { t.Fatal(err) }
	b, err = NormalizeIdentity(ObservationIP, "2001:db8::1", nil, "", "")
	if err != nil { t.Fatal(err) }
	if a != b { t.Fatalf("IP identity differs: %s != %s", a, b) }
}

func TestAssignIDsIsIdempotentAndPreservesConflicts(t *testing.T) {
	a := validObservation()
	if err := a.AssignIDs(); err != nil { t.Fatal(err) }
	b := validObservation()
	if err := b.AssignIDs(); err != nil { t.Fatal(err) }
	if a.IdentityKey != b.IdentityKey || a.ObservationID != b.ObservationID { t.Fatal("identical replay must retain stable IDs") }
	c := validObservation()
	c.Value = "nginx"
	if err := c.AssignIDs(); err != nil { t.Fatal(err) }
	if c.IdentityKey != a.IdentityKey { t.Fatal("conflicting claim should share entity identity") }
	if c.ObservationID == a.ObservationID { t.Fatal("conflicting claim must remain a distinct observation") }
	d := validObservation()
	d.SourceRunID = "run-456"
	if err := d.AssignIDs(); err != nil { t.Fatal(err) }
	if d.IdentityKey != a.IdentityKey || d.ObservationID == a.ObservationID { t.Fatal("separate runs must retain separate source observations") }
}

func TestPortObservationRequiresValidPort(t *testing.T) {
	o := validObservation()
	o.Type, o.Target = ObservationPort, "192.0.2.10"
	if err := o.Validate(); err == nil { t.Fatal("expected missing port error") }
	p, err := ParsePort("443")
	if err != nil { t.Fatal(err) }
	o.Port = p
	if err := o.Validate(); err != nil { t.Fatal(err) }
	if _, err := ParsePort("0"); err == nil { t.Fatal("port zero accepted") }
	if _, err := ParsePort("65536"); err == nil { t.Fatal("out-of-range port accepted") }
}
