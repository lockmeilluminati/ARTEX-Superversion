// Package assetgraph derives explainable relationships between existing ARTEX
// asset records. It does not create or mutate assets and does not treat inferred
// relationships as verified facts.
package assetgraph

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"

	"github.com/Autumn-27/artex/db"
)

// Edge is a deterministic, read-only relationship projection between two assets.
// Evidence describes the exact source fields that caused the relationship.
type Edge struct {
	FromAssetID int64  `json:"from_asset_id"`
	ToAssetID   int64  `json:"to_asset_id"`
	Relation    string `json:"relation"`
	Evidence    string `json:"evidence"`
	Confidence  string `json:"confidence"`
	Verification string `json:"verification"`
}

type edgeKey struct { from, to int64; relation string }

// Correlate derives only relationships supported by explicit asset fields.
// Shared company, task membership, subnet proximity and technology labels are
// deliberately not sufficient to create an edge.
func Correlate(assets []db.Asset) []Edge {
	byID := make(map[int64]db.Asset, len(assets))
	domains := map[string]int64{}
	ips := map[string]int64{}
	services := make(map[string]int64)
	for _, a := range assets {
		byID[a.ID] = a
		switch a.Type {
		case "root_domain", "subdomain":
			if k := domainKey(a.Domain); k != "" {
				domains[k] = a.ID
			}
		case "ip":
			if k := ipKey(a.IP); k != "" { ips[k] = a.ID }
		case "service":
			if k := originKey(a.URL); k != "" { services[k] = a.ID }
		}
	}
	edges := map[edgeKey]Edge{}
	add := func(from, to int64, rel, evidence, confidence, verification string) {
		if from <= 0 || to <= 0 || from == to { return }
		k := edgeKey{from, to, rel}
		if _, ok := edges[k]; !ok {
			edges[k] = Edge{FromAssetID: from, ToAssetID: to, Relation: rel, Evidence: evidence, Confidence: confidence, Verification: verification}
		}
	}
	for _, a := range assets {
		switch a.Type {
		case "subdomain":
			if root := domainKey(a.RootDomain); root != "" {
				if id := domains[root]; id != 0 {
					add(a.ID, id, "belongs_to_root_domain", "assets.root_domain matches root-domain asset.domain", "high", "observed")
				}
			}
			candidates := append([]string(nil), a.RecordValue...)
			if a.IP != "" { candidates = append(candidates, a.IP) }
			for _, value := range candidates {
				if ip := ipKey(value); ip != "" {
					if id := ips[ip]; id != 0 {
						add(a.ID, id, "resolves_to_ip", "subdomain A/AAAA record value or ip field matches IP asset", "high", "observed")
					}
				}
			}
		case "ip":
			for _, domain := range a.BoundDomains {
				if id := domains[domainKey(domain)]; id != 0 {
					add(id, a.ID, "resolves_to_ip", "IP asset.bound_domains contains domain", "high", "observed")
				}
			}
		case "service", "endpoint":
			host := hostKey(a.URL)
			if host == "" { host = domainKey(a.Domain) }
			if host != "" {
				if id := domains[host]; id != 0 {
					add(a.ID, id, "hosted_on_domain", "URL host/domain matches domain asset", "high", "observed")
				}
				if id := ips[host]; id != 0 {
					add(a.ID, id, "hosted_on_ip", "URL host matches IP asset", "high", "observed")
				}
			}
			if a.Type == "endpoint" {
				if id := services[originKey(a.URL)]; id != 0 {
					add(a.ID, id, "endpoint_of_service", "endpoint URL origin matches HTTP service URL origin", "high", "inferred")
				}
			}
		}
	}
	out := make([]Edge, 0, len(edges))
	for _, e := range edges { out = append(out, e) }
	sort.Slice(out, func(i, j int) bool {
		if out[i].FromAssetID != out[j].FromAssetID { return out[i].FromAssetID < out[j].FromAssetID }
		if out[i].ToAssetID != out[j].ToAssetID { return out[i].ToAssetID < out[j].ToAssetID }
		return out[i].Relation < out[j].Relation
	})
	return out
}

func domainKey(s string) string { return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(s)), ".") }
func ipKey(s string) string {
	s = strings.TrimSpace(s)
	if ip := net.ParseIP(s); ip != nil { return ip.String() }
	return ""
}
func hostKey(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil { return "" }
	host := strings.ToLower(u.Hostname())
	if host == "" { return "" }
	if ip := net.ParseIP(host); ip != nil { return ip.String() }
	return domainKey(host)
}
func originKey(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" { return "" }
	return fmt.Sprintf("%s://%s", strings.ToLower(u.Scheme), strings.ToLower(u.Host))
}

// AssetIDs returns the unique IDs referenced by a correlation result.
func AssetIDs(edges []Edge) []int64 {
	set := map[int64]bool{}
	for _, e := range edges { set[e.FromAssetID] = true; set[e.ToAssetID] = true }
	out := make([]int64, 0, len(set))
	for id := range set { out = append(out, id) }
	sort.Slice(out, func(i,j int) bool { return out[i] < out[j] })
	return out
}
