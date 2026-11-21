package db

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Autumn-27/artex/normalizer"
)

func TestNewCollectionRunIDIsUUID(t *testing.T) {
	id, err := NewCollectionRunID()
	if err != nil { t.Fatal(err) }
	if !uuidPattern.MatchString(id) { t.Fatalf("invalid UUID %q", id) }
	other, err := NewCollectionRunID()
	if err != nil { t.Fatal(err) }
	if id == other { t.Fatal("two collection runs received the same ID") }
}

// PG: exact observation replay is a no-op, same-ID payload mutation is a
// conflict, and a different run can preserve a contradictory claim.
func TestObservationClaimPersistenceAndReplay(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil { SkipPGUnavailable(t, err) }
	t.Cleanup(func() { d.Close() })
	task, err := d.CreateTaskWithOptions("observation provenance", "test", TaskCreateOptions{})
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { cleanupErr(t, "task", d.DeleteTask(task.ID)) })

	ctx := context.Background()
	runID, err := d.CreateCollectionRun(ctx, CollectionRun{
		TaskID: task.ID, ExplorationID: &task.ExplorationID,
		SourceToolID: "external.nmap", ExecutionMode: "cli", Status: "running",
	})
	if err != nil { t.Fatal(err) }
	digest := strings.Repeat("a", 64)
	evidenceID, err := NewCollectionRunID()
	if err != nil { t.Fatal(err) }
	if err := d.RegisterEvidenceArtifactMetadata(ctx, EvidenceArtifactMetadata{
		ID: evidenceID, SHA256: digest, ByteLength: 12,
		MediaType: "application/xml", StorageKey: "sha256/" + digest,
		TaskID: &task.ID, CollectionRunID: &runID, SourceToolID: "external.nmap",
		OriginalName: "scan.xml",
	}); err != nil { t.Fatal(err) }

	now := time.Now().UTC().Truncate(time.Microsecond)
	o := normalizer.Observation{
		Type: normalizer.ObservationIP, Target: "192.0.2.10",
		SourceToolID: "external.nmap", SourceRunID: runID, TaskID: task.ID,
		ObservedAt: now, Scope: normalizer.ScopeInScope,
		Confidence: normalizer.ConfidenceHigh, Verification: normalizer.VerificationUnverified,
		Collection: normalizer.CollectionTestedPositive,
		Evidence: []normalizer.EvidenceReference{{Kind: "artifact", URI: evidenceID, Digest: digest, MediaType: "application/xml", ObservedAt: now}},
	}
	if err := o.AssignIDs(); err != nil { t.Fatal(err) }
	inserted, err := d.InsertObservationClaim(ctx, o, evidenceID, ObservationClaimLink{})
	if err != nil || !inserted { t.Fatalf("first insert=%v err=%v", inserted, err) }
	inserted, err = d.InsertObservationClaim(ctx, o, evidenceID, ObservationClaimLink{})
	if err != nil || inserted { t.Fatalf("replay insert=%v err=%v", inserted, err) }

	mutated := o
	mutated.Value = "changed payload"
	if _, err := d.InsertObservationClaim(ctx, mutated, evidenceID, ObservationClaimLink{}); err == nil || !strings.Contains(err.Error(), "collision") {
		t.Fatalf("same-ID payload mutation error=%v, want collision", err)
	}

	run2, err := d.CreateCollectionRun(ctx, CollectionRun{
		TaskID: task.ID, ExplorationID: &task.ExplorationID,
		SourceToolID: "external.nmap", ExecutionMode: "cli", Status: "running",
	})
	if err != nil { t.Fatal(err) }
	evidence2, err := NewCollectionRunID()
	if err != nil { t.Fatal(err) }
	if err := d.RegisterEvidenceArtifactMetadata(ctx, EvidenceArtifactMetadata{
		ID: evidence2, SHA256: strings.Repeat("b", 64), ByteLength: 13,
		MediaType: "application/xml", StorageKey: "sha256/" + strings.Repeat("b", 64),
		TaskID: &task.ID, CollectionRunID: &run2, SourceToolID: "external.nmap",
	}); err != nil { t.Fatal(err) }
	conflict := o
	conflict.SourceRunID = run2
	conflict.Evidence = []normalizer.EvidenceReference{{Kind: "artifact", URI: evidence2, Digest: strings.Repeat("b", 64), MediaType: "application/xml", ObservedAt: now.Add(time.Second)}}
	conflict.ObservedAt = now.Add(time.Second)
	conflict.Value = "different source claim"
	if err := conflict.AssignIDs(); err != nil { t.Fatal(err) }
	inserted, err = d.InsertObservationClaim(ctx, conflict, evidence2, ObservationClaimLink{})
	if err != nil || !inserted { t.Fatalf("contradictory claim insert=%v err=%v", inserted, err) }

	var count int
	if err := d.QueryRowContext(ctx, `SELECT count(*) FROM observation_claims WHERE identity_key=$1`, o.IdentityKey).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 { t.Fatalf("claims with shared identity=%d, want 2", count) }
}
