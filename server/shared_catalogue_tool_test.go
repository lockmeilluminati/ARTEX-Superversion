package server

import "testing"

func TestCatalogueMatches(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		fields []string
		want   bool
	}{
		{name: "empty query matches all", query: "", fields: []string{"Nmap"}, want: true},
		{name: "case insensitive match", query: "nUcLeI", fields: []string{"Nuclei scanner"}, want: true},
		{name: "searches across fields", query: "web security", fields: []string{"Yakit", "Web security"}, want: true},
		{name: "no match", query: "forensics", fields: []string{"Nmap", "Network analysis"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := catalogueMatches(tt.query, tt.fields...); got != tt.want {
				t.Fatalf("catalogueMatches(%q, %v) = %v, want %v", tt.query, tt.fields, got, tt.want)
			}
		})
	}
}
