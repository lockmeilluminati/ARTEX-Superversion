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
	if _, err := d.CreateCollectionRun(ctx, CollectionRun{
		TaskID: task.ID, SourceToolID: "external.nmap", Status: "running",
	}); err == nil {
		t.Fatal("new collection run bypassed the planned state")
	}
	exitCode := 0
	if _, err := d.CreateCollectionRun(ctx, CollectionRun{
		TaskID: task.ID, SourceToolID: "external.nmap", Status: "planned", ExitCode: &exitCode,
	}); err == nil {
		t.Fatal("new collection run accepted terminal execution fields")
	}
	runID, err := d.CreateCollectionRun(ctx, CollectionRun{
		TaskID: task.ID, ExplorationID: &task.ExplorationID,
		SourceToolID: "external.nmap", ExecutionMode: "cli", Status: "planned",
	})
	if err != nil { t.Fatal(err) }
	if err := d.StartCollectionRun(ctx, runID, time.Now().UTC()); err != nil { t.Fatal(err) }
	digest := strings.Repeat("a", 64)
	evidenceID, err := NewCollectionRunID()
	if err != nil { t.Fatal(err) }
	if err := d.RegisterEvidenceArtifactMetadata(ctx, EvidenceArtifactMetadata{
		ID: evidenceID, SHA256: digest, ByteLength: 12,
		MediaType: "application/xml", StorageKey: "sha256/" + digest,
		TaskID: &task.ID, CollectionRunID: &runID, SourceToolID: "external.nmap",
		OriginalName: "scan.xml",
	}); err != nil { t.Fatal(err) }
	// An artifact cannot be attached to a run under a different tool identity.
	mismatched := EvidenceArtifactMetadata{
		ID: evidenceID, SHA256: digest, ByteLength: 12,
		MediaType: "application/xml", StorageKey: "sha256/" + digest + "-mismatch",
		TaskID: &task.ID, CollectionRunID: &runID, SourceToolID: "external.nuclei",
		OriginalName: "scan.xml",
	}
	if err := d.RegisterEvidenceArtifactMetadata(ctx, mismatched); err == nil || !strings.Contains(err.Error(), "do not match") {
		t.Fatalf("mismatched source tool registration error=%v, want run mismatch", err)
	}

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

	// A claim cannot borrow a valid artifact while asserting another tool or task.
	wrongTool := o
	wrongTool.SourceToolID = "external.nuclei"
	if err := wrongTool.AssignIDs(); err != nil { t.Fatal(err) }
	if _, err := d.InsertObservationClaim(ctx, wrongTool, evidenceID, ObservationClaimLink{}); err == nil {
		t.Fatal("claim with mismatched source tool was accepted")
	}
	wrongTask := o
	wrongTask.TaskID = task.ID + 999
	if err := wrongTask.AssignIDs(); err != nil { t.Fatal(err) }
	if _, err := d.InsertObservationClaim(ctx, wrongTask, evidenceID, ObservationClaimLink{}); err == nil {
		t.Fatal("claim with mismatched task was accepted")
	}

	inserted, err = d.InsertObservationClaim(ctx, o, evidenceID, ObservationClaimLink{})
	if err != nil || inserted { t.Fatalf("replay insert=%v err=%v", inserted, err) }

	mutated := o
	mutated.Value = "changed payload"
	if _, err := d.InsertObservationClaim(ctx, mutated, evidenceID, ObservationClaimLink{}); err == nil || !strings.Contains(err.Error(), "collision") {
		t.Fatalf("same-ID payload mutation error=%v, want collision", err)
	}

	run2, err := d.CreateCollectionRun(ctx, CollectionRun{
		TaskID: task.ID, ExplorationID: &task.ExplorationID,
		SourceToolID: "external.nmap", ExecutionMode: "cli", Status: "planned",
	})
	if err != nil { t.Fatal(err) }
	if err := d.StartCollectionRun(ctx, run2, time.Now().UTC()); err != nil { t.Fatal(err) }
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

func TestCollectionRunLifecycleTransitions(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil { SkipPGUnavailable(t, err) }
	t.Cleanup(func() { d.Close() })
	task, err := d.CreateTaskWithOptions("collection run lifecycle", "test", TaskCreateOptions{})
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { cleanupErr(t, "task", d.DeleteTask(task.ID)) })

	ctx := context.Background()
	runID, err := d.CreateCollectionRun(ctx, CollectionRun{
		TaskID: task.ID, ExplorationID: &task.ExplorationID,
		SourceToolID: "external.nmap", ExecutionMode: "cli", Status: "planned",
	})
	if err != nil { t.Fatal(err) }
	var startIsNull bool
	if err := d.QueryRowContext(ctx, `SELECT started_at IS NULL FROM collection_runs WHERE id=$1::uuid`, runID).Scan(&startIsNull); err != nil {
		t.Fatal(err)
	}
	if !startIsNull {
		t.Fatal("planned collection run has a start timestamp before process launch")
	}
	if err := d.FinishCollectionRun(ctx, runID, "succeeded", time.Now().UTC(), nil, "premature.xml"); err == nil {
		t.Fatal("planned collection run was finalized before process start")
	}

	started := time.Now().UTC().Truncate(time.Microsecond)
	if err := d.StartCollectionRun(ctx, runID, started); err != nil {
		t.Fatalf("start planned run: %v", err)
	}
	var status string
	var storedStart time.Time
	if err := d.QueryRowContext(ctx, `SELECT status, started_at FROM collection_runs WHERE id=$1::uuid`, runID).Scan(&status, &storedStart); err != nil {
		t.Fatal(err)
	}
	if status != "running" || !storedStart.Equal(started) {
		t.Fatalf("run state=(%q,%s), want (running,%s)", status, storedStart, started)
	}
	if err := d.StartCollectionRun(ctx, runID, started.Add(time.Second)); err == nil {
		t.Fatal("duplicate process start was accepted")
	}
	exitCode := 0
	finished := started.Add(time.Minute)
	if err := d.FinishCollectionRun(ctx, runID, "succeeded", finished, &exitCode, "output.xml"); err != nil {
		t.Fatalf("finish running run: %v", err)
	}
	if err := d.FinishCollectionRun(ctx, runID, "failed", finished.Add(time.Second), nil, "replacement.xml"); err == nil {
		t.Fatal("terminal collection run was finalized a second time")
	}
	var finalStatus, outputLocator string
	var finalFinished time.Time
	if err := d.QueryRowContext(ctx, `SELECT status, output_locator, finished_at FROM collection_runs WHERE id=$1::uuid`, runID).Scan(&finalStatus, &outputLocator, &finalFinished); err != nil {
		t.Fatal(err)
	}
	if finalStatus != "succeeded" || outputLocator != "output.xml" || !finalFinished.Equal(finished) {
		t.Fatalf("second finalization changed terminal record: status=%q output=%q finished=%s", finalStatus, outputLocator, finalFinished)
	}
	if err := d.StartCollectionRun(ctx, runID, started.Add(2*time.Minute)); err == nil {
		t.Fatal("terminal collection run was restarted")
	}
}
