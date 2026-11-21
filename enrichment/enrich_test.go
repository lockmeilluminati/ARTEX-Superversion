package enrichment

import (
	"testing"

	"github.com/Autumn-27/artex/db"
)

func TestDeriveUsesExplicitEvidenceAndLabelsPortInference(t *testing.T) {
	port := 443
	status := 200
	assets := []db.Asset{
		{ID: 1, Type: "service", URL: "https://example.test", ServiceName: "nginx", ServiceType: "https", StatusCode: &status, PageTitle: "Portal", Technologies: []string{"nginx/1.25"}},
		{ID: 2, Type: "service", Port: &port},
	}
	got := Derive(assets)
	if len(got.Claims) != 6 {
		t.Fatalf("claims=%+v", got.Claims)
	}
	foundObserved, foundInferred := false, false
	for _, c := range got.Claims {
		if c.AssetID == 1 && c.Kind == "technology" && c.Value == "nginx/1.25" && c.Verification == "observed" && c.Confidence == "medium" {
			foundObserved = true
		}
		if c.AssetID == 2 && c.Kind == "possible_service" && c.Value == "https" && c.Verification == "inferred" && c.Confidence == "low" {
			foundInferred = true
		}
	}
	if !foundObserved || !foundInferred {
		t.Fatalf("expected explicit technology and cautious port inference: %+v", got.Claims)
	}
}

func TestDeriveDoesNotInventServiceFromUnknownPort(t *testing.T) {
	port := 49152
	got := Derive([]db.Asset{{ID: 3, Type: "service", Port: &port}})
	if len(got.Claims) != 0 {
		t.Fatalf("unexpected claims: %+v", got.Claims)
	}
}

func TestDeriveIncludesExplainableAssetRelationships(t *testing.T) {
	assets := []db.Asset{
		{ID: 10, Type: "root_domain", Domain: "example.test"},
		{ID: 11, Type: "subdomain", Domain: "www.example.test", RootDomain: "example.test"},
	}
	got := Derive(assets)
	if len(got.Relationships) != 1 {
		t.Fatalf("relationships=%+v", got.Relationships)
	}
	r := got.Relationships[0]
	if r.FromAssetID != 11 || r.ToAssetID != 10 || r.Relation != "belongs_to_root_domain" || r.Evidence == "" || r.Confidence != "high" || r.Verification != "observed" {
		t.Fatalf("unexpected relationship: %+v", r)
	}
}
