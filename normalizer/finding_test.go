package normalizer

import (
	"testing"
	"time"
)

func TestNormalizePreservesSourceAndTarget(t *testing.T) {
	now := time.Now().UTC()
	got := Normalize(ExternalFinding{SourceToolID:"scanner-x", SourceToolVersion:"1.2", Title:"Finding", Target:"lab.local", ObservedAt:now})
	if got.SourceToolID != "scanner-x" || got.Asset != "lab.local" || got.SourceToolVersion != "1.2" {
		t.Fatalf("unexpected normalized finding: %#v", got)
	}
}
