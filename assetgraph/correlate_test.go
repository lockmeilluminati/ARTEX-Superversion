package assetgraph

import (
	"reflect"
	"testing"

	"github.com/Autumn-27/artex/db"
)

func TestCorrelateBuildsExplainableConservativeEdges(t *testing.T) {
	port := 443
	assets := []db.Asset{
		{ID: 1, Type: "root_domain", Domain: "example.com"},
		{ID: 2, Type: "subdomain", Domain: "www.example.com", RootDomain: "example.com", RecordType: "A", RecordValue: []string{"192.0.2.10"}},
		{ID: 3, Type: "ip", IP: "192.0.2.10", BoundDomains: []string{"www.example.com"}},
		{ID: 4, Type: "service", URL: "https://www.example.com", Domain: "www.example.com", Port: &port, Technologies: []string{"nginx"}},
		{ID: 5, Type: "endpoint", URL: "https://www.example.com/login", Method: "GET"},
		{ID: 6, Type: "subdomain", Domain: "unrelated.example.net", RootDomain: "example.net", CompanyID: int64Ptr(7)},
	}
	got := Correlate(assets)
	want := []Edge{
		{FromAssetID: 2, ToAssetID: 1, Relation: "belongs_to_root_domain", Evidence: "assets.root_domain matches root-domain asset.domain", Confidence: "high", Verification: "observed"},
		{FromAssetID: 2, ToAssetID: 3, Relation: "resolves_to_ip", Evidence: "subdomain A/AAAA record value or ip field matches IP asset", Confidence: "high", Verification: "observed"},
		{FromAssetID: 2, ToAssetID: 3, Relation: "resolves_to_ip", Evidence: "subdomain A/AAAA record value or ip field matches IP asset", Confidence: "high", Verification: "observed"},
		{FromAssetID: 4, ToAssetID: 2, Relation: "hosted_on_domain", Evidence: "URL host/domain matches domain asset", Confidence: "high", Verification: "observed"},
		{FromAssetID: 5, ToAssetID: 2, Relation: "hosted_on_domain", Evidence: "URL host/domain matches domain asset", Confidence: "high", Verification: "observed"},
		{FromAssetID: 5, ToAssetID: 4, Relation: "endpoint_of_service", Evidence: "endpoint URL origin matches HTTP service URL origin", Confidence: "high", Verification: "inferred"},
	}
	// The IP->domain reverse observation and subdomain->IP direct observation
	// collapse to one deterministic relationship.
	if len(got) != 5 { t.Fatalf("got %d edges: %+v", len(got), got) }
	for _, e := range got {
		if e.Confidence == "" || e.Verification == "" || e.Evidence == "" { t.Fatalf("edge lacks rationale: %+v", e) }
	}
	if got[0].Relation != "belongs_to_root_domain" || got[1].Relation != "resolves_to_ip" || got[2].Relation != "hosted_on_domain" || got[3].Relation != "hosted_on_domain" || got[4].Relation != "endpoint_of_service" {
		t.Fatalf("unexpected deterministic edges: %+v", got)
	}
	again := Correlate(assets)
	if !reflect.DeepEqual(got, again) { t.Fatal("correlation output is not deterministic") }
	if ids := AssetIDs(got); !reflect.DeepEqual(ids, []int64{1,2,3,4,5}) { t.Fatalf("asset IDs = %v", ids) }
}

func TestCorrelateDoesNotInferFromCompanyOrSubnetAlone(t *testing.T) {
	assets := []db.Asset{
		{ID: 1, Type: "ip", IP: "192.0.2.1", CompanyID: int64Ptr(9)},
		{ID: 2, Type: "ip", IP: "192.0.2.2", CompanyID: int64Ptr(9)},
		{ID: 3, Type: "subdomain", Domain: "x.example.com", CompanyID: int64Ptr(9)},
	}
	if got := Correlate(assets); len(got) != 0 { t.Fatalf("unsupported edges inferred: %+v", got) }
}
func int64Ptr(v int64) *int64 { return &v }
