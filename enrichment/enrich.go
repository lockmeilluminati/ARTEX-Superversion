// Package enrichment derives explainable, non-mutating context from existing ARTEX assets.
package enrichment

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Autumn-27/artex/assetgraph"
	"github.com/Autumn-27/artex/db"
)

// Claim is a source-attributed enrichment projection. It is not a verified fact
// and is deliberately not persisted independently of ARTEX's existing stores.
type Claim struct {
	AssetID      int64    `json:"asset_id"`
	Kind         string   `json:"kind"`
	Value        string   `json:"value"`
	SourceFields []string `json:"source_fields"`
	Evidence     string   `json:"evidence"`
	Confidence   string   `json:"confidence"`
	Verification string   `json:"verification"`
	Rationale    string   `json:"rationale"`
}

// Relationship retains the assetgraph edge's evidence and confidence labels.
type Relationship struct {
	FromAssetID  int64  `json:"from_asset_id"`
	ToAssetID    int64  `json:"to_asset_id"`
	Relation     string `json:"relation"`
	Evidence     string `json:"evidence"`
	Confidence   string `json:"confidence"`
	Verification string `json:"verification"`
}

// Result is a read-only projection. Persist source observations in ARTEX's
// existing asset/fact/evidence stores; do not treat these candidates as facts.
type Result struct {
	Claims        []Claim        `json:"claims"`
	Relationships []Relationship `json:"relationships"`
}

func Derive(assets []db.Asset) Result {
	out := Result{Claims: make([]Claim, 0), Relationships: make([]Relationship, 0)}
	for _, a := range assets {
		add := func(kind, value string, fields []string, evidence, confidence, verification, rationale string) {
			value = strings.TrimSpace(value)
			if a.ID <= 0 || value == "" {
				return
			}
			out.Claims = append(out.Claims, Claim{AssetID: a.ID, Kind: kind, Value: value, SourceFields: fields, Evidence: evidence, Confidence: confidence, Verification: verification, Rationale: rationale})
		}
		if a.ServiceName != "" {
			add("service_identity", a.ServiceName, []string{"assets.service_name"}, "ARTEX asset service_name field", "high", "observed", "Service name was explicitly recorded on the asset; this projection does not independently verify the service.")
		}
		if a.ServiceType != "" {
			add("service_type", a.ServiceType, []string{"assets.service_type"}, "ARTEX asset service_type field", "high", "observed", "Service type was explicitly recorded on the asset; source execution evidence remains authoritative.")
		}
		if a.PageTitle != "" {
			add("http_page_title", a.PageTitle, []string{"assets.page_title"}, "ARTEX asset page_title field", "medium", "observed", "Page title is observed HTTP metadata, not proof of the application's identity.")
		}
		if a.StatusCode != nil {
			add("http_status", strconv.Itoa(*a.StatusCode), []string{"assets.status_code"}, "ARTEX asset status_code field", "high", "observed", "HTTP status was explicitly recorded; it may change between requests.")
		}
		for _, tech := range a.Technologies {
			add("technology", tech, []string{"assets.technologies"}, "ARTEX asset technologies field", "medium", "observed", "Technology label was explicitly recorded by an upstream collector; version and fingerprint accuracy are not independently verified.")
		}
		for _, item := range a.OpenPorts {
			port := integer(item["port"])
			service, _ := item["service"].(string)
			if port > 0 && strings.TrimSpace(service) != "" {
				add("service_identity", service, []string{"assets.open_ports[].port", "assets.open_ports[].service"}, fmt.Sprintf("open port %d explicitly carries service label %q", port, strings.TrimSpace(service)), "medium", "observed", "The collector supplied a service label for this open port; ARTEX has not independently verified the protocol.")
			}
		}
		if a.Port != nil && *a.Port > 0 {
			if name, ok := commonPortHint(*a.Port); ok {
				add("possible_service", name, []string{"assets.port"}, fmt.Sprintf("port number %d", *a.Port), "low", "inferred", "This is only a conventional port association. A port number alone does not establish which service is running.")
			}
		}
	}
	for _, edge := range assetgraph.Correlate(assets) {
		out.Relationships = append(out.Relationships, Relationship{
			FromAssetID: edge.FromAssetID, ToAssetID: edge.ToAssetID, Relation: edge.Relation,
			Evidence: edge.Evidence, Confidence: edge.Confidence, Verification: edge.Verification,
		})
	}
	sort.Slice(out.Claims, func(i, j int) bool {
		a, b := out.Claims[i], out.Claims[j]
		if a.AssetID != b.AssetID {
			return a.AssetID < b.AssetID
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Value < b.Value
	})
	return out
}

func integer(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(n))
		return i
	default:
		return 0
	}
}

func commonPortHint(port int) (string, bool) {
	switch port {
	case 21:
		return "ftp", true
	case 22:
		return "ssh", true
	case 25, 587:
		return "smtp", true
	case 53:
		return "dns", true
	case 80, 8080, 8000:
		return "http", true
	case 110:
		return "pop3", true
	case 143:
		return "imap", true
	case 443, 8443:
		return "https", true
	case 3306:
		return "mysql", true
	case 5432:
		return "postgresql", true
	case 6379:
		return "redis", true
	case 27017:
		return "mongodb", true
	default:
		return "", false
	}
}
