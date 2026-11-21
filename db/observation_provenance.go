package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Autumn-27/artex/normalizer"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// NewCollectionRunID creates the canonical process-execution identifier. Call
// this at the execution boundary before launching the external process, then
// pass the same ID through output capture and ingestion. A retry gets a new ID.
func NewCollectionRunID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate collection run id: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	s := hex.EncodeToString(b[:])
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:], nil
}

type CollectionRun struct {
	ID                string
	TaskID            int64
	ExplorationID     *int64
	SourceToolID      string
	SourceToolVersion string
	ExecutionMode     string
	Platform          string
	Status            string
	StartedAt         time.Time
	FinishedAt        *time.Time
	ExitCode          *int
	OutputLocator     string
	Metadata          json.RawMessage
}

// CreateCollectionRun persists a run before process launch. It deliberately
// does not infer a run ID from an ingestion attempt or artifact filename.
func (d *DB) CreateCollectionRun(ctx context.Context, run CollectionRun) (string, error) {
	if run.ID == "" {
		id, err := NewCollectionRunID()
		if err != nil { return "", err }
		run.ID = id
	}
	if !uuidPattern.MatchString(run.ID) { return "", errors.New("collection run id must be a UUID") }
	if run.TaskID <= 0 { return "", errors.New("collection run task id must be positive") }
	if strings.TrimSpace(run.SourceToolID) == "" { return "", errors.New("source tool id is required") }
	if run.ExecutionMode == "" { run.ExecutionMode = "unknown" }
	if run.Status == "" { run.Status = "planned" }
	if run.Status != "planned" {
		return "", errors.New("new collection run must start in planned status")
	}
	if run.FinishedAt != nil || run.ExitCode != nil {
		return "", errors.New("new collection run cannot have terminal execution fields")
	}
	if run.StartedAt.IsZero() { run.StartedAt = time.Now().UTC() }
	if len(run.Metadata) == 0 { run.Metadata = json.RawMessage(`{}`) }
	if !json.Valid(run.Metadata) || len(run.Metadata) > 64*1024 { return "", errors.New("collection run metadata must be valid JSON under 64 KiB") }

	_, err := d.ExecContext(ctx, `
INSERT INTO collection_runs
(id, task_id, exploration_id, source_tool_id, source_tool_version, execution_mode, platform, status, started_at, output_locator, metadata)
VALUES ($1::uuid,$2,$3,$4,NULLIF($5,''),$6,$7,$8,$9,$10,$11::jsonb)`,
		run.ID, run.TaskID, run.ExplorationID, run.SourceToolID, run.SourceToolVersion,
		run.ExecutionMode, run.Platform, run.Status, run.StartedAt, run.OutputLocator, string(run.Metadata))
	if err != nil { return "", fmt.Errorf("create collection run: %w", err) }
	return run.ID, nil
}

// StartCollectionRun records that the external process was actually launched.
// It is a one-way transition: duplicate starts and starts after a terminal state
// are rejected so callers cannot silently rewrite execution history.
func (d *DB) StartCollectionRun(ctx context.Context, id string, started time.Time) error {
	if !uuidPattern.MatchString(id) {
		return errors.New("collection run id must be a UUID")
	}
	if started.IsZero() {
		started = time.Now().UTC()
	}
	res, err := d.ExecContext(ctx, `
UPDATE collection_runs
SET status='running', started_at=$2
WHERE id=$1::uuid AND status='planned'`, id, started)
	if err != nil {
		return fmt.Errorf("start collection run: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) FinishCollectionRun(ctx context.Context, id, status string, finished time.Time, exitCode *int, outputLocator string) error {
	if !uuidPattern.MatchString(id) { return errors.New("collection run id must be a UUID") }
	switch status {
	case "succeeded", "partial", "failed", "cancelled":
	default: return fmt.Errorf("invalid terminal collection run status %q", status)
	}
	if finished.IsZero() { finished = time.Now().UTC() }
	res, err := d.ExecContext(ctx, `
UPDATE collection_runs SET status=$2, finished_at=$3, exit_code=$4, output_locator=$5
WHERE id=$1::uuid AND status='running'`, id, status, finished, exitCode, outputLocator)
	if err != nil { return fmt.Errorf("finish collection run: %w", err) }
	n, err := res.RowsAffected()
	if err != nil { return err }
	if n == 0 { return sql.ErrNoRows }
	return nil
}

type EvidenceArtifactMetadata struct {
	ID             string
	SHA256         string
	ByteLength     int64
	MediaType      string
	StorageKey     string
	TaskID         *int64
	CollectionRunID *string
	SourceToolID   string
	OriginalName   string
	CapturedAt     time.Time
}

// RegisterEvidenceArtifactMetadata records metadata only after the evidence
// service has durably written the immutable blob and independently computed
// its digest/length. This method cannot register or verify file bytes itself.
func (d *DB) RegisterEvidenceArtifactMetadata(ctx context.Context, a EvidenceArtifactMetadata) error {
	if a.ID == "" {
		id, err := NewCollectionRunID()
		if err != nil { return err }
		a.ID = id
	}
	if !uuidPattern.MatchString(a.ID) { return errors.New("evidence artifact id must be a UUID") }
	if !sha256Pattern.MatchString(a.SHA256) { return errors.New("evidence artifact sha256 must be lowercase hex") }
	if a.ByteLength < 0 { return errors.New("evidence artifact byte length cannot be negative") }
	if strings.TrimSpace(a.StorageKey) == "" || len(a.StorageKey) > 512 { return errors.New("opaque evidence storage key is required") }
	if len(a.OriginalName) > 512 { return errors.New("original artifact name exceeds 512 bytes") }
	if a.CapturedAt.IsZero() { a.CapturedAt = time.Now().UTC() }
	if a.MediaType == "" { a.MediaType = "application/octet-stream" }
	if a.TaskID == nil || *a.TaskID <= 0 || a.CollectionRunID == nil || !uuidPattern.MatchString(*a.CollectionRunID) {
		return errors.New("evidence artifact requires a task and collection run")
	}
	if strings.TrimSpace(a.SourceToolID) == "" { return errors.New("evidence artifact source tool id is required") }
	res, err := d.ExecContext(ctx, `
INSERT INTO evidence_artifacts
(id, sha256, byte_length, media_type, storage_key, task_id, collection_run_id, source_tool_id, original_name, captured_at)
SELECT $1::uuid,$2,$3,$4,$5,$6,$7::uuid,$8,$9,$10
WHERE EXISTS (
 SELECT 1 FROM collection_runs cr
 WHERE cr.id=$7::uuid AND cr.task_id=$6 AND cr.source_tool_id=$8
)`,
		a.ID, a.SHA256, a.ByteLength, a.MediaType, a.StorageKey, a.TaskID, *a.CollectionRunID, a.SourceToolID, a.OriginalName, a.CapturedAt)
	if err != nil { return fmt.Errorf("register evidence artifact metadata: %w", err) }
	n, err := res.RowsAffected()
	if err != nil { return err }
	if n == 0 { return errors.New("evidence artifact task, source tool and collection run do not match an existing run") }
	return nil
}

func nullableUUID(v *string) any {
	if v == nil || *v == "" { return nil }
	return *v
}

type ObservationClaimLink struct {
	AssetID            *int64
	FindingID          *int64
	ExplorationNodeID  *int64
}

// InsertObservationClaim appends one source claim. Exact replay of an identical
// payload is a no-op; reuse of an observation ID for different content is an
// error. Identity keys are intentionally non-unique so contradictory claims
// from different runs remain visible.
func (d *DB) InsertObservationClaim(ctx context.Context, o normalizer.Observation, evidenceID string, link ObservationClaimLink) (inserted bool, err error) {
	if err := o.Validate(); err != nil { return false, fmt.Errorf("validate observation: %w", err) }
	if o.ObservationID == "" || o.IdentityKey == "" {
		if err := o.AssignIDs(); err != nil { return false, err }
	}
	if !uuidPattern.MatchString(o.SourceRunID) { return false, errors.New("durable observation source_run_id must be the actual collection run UUID") }
	if !uuidPattern.MatchString(evidenceID) { return false, errors.New("durable observation requires a registered evidence artifact UUID") }
	if len(o.Evidence) == 0 || !sha256Pattern.MatchString(o.Evidence[0].Digest) {
		return false, errors.New("durable observation requires a verified evidence SHA-256 reference")
	}
	if len(o.Attributes) > 128 { return false, errors.New("observation has too many attributes") }
	payload, err := json.Marshal(o)
	if err != nil { return false, err }
	if len(payload) > 256*1024 { return false, errors.New("observation payload exceeds 256 KiB") }

	var port any
	if o.Port != nil { port = int(*o.Port) }
	var version any
	if o.SourceToolVersion != "" { version = o.SourceToolVersion }
	res, err := d.ExecContext(ctx, `
INSERT INTO observation_claims
(observation_id, identity_key, observation_type, target, port, protocol, name, value,
 source_tool_id, source_tool_version, source_run_id, task_id, observed_at,
 evidence_id, evidence_sha256, scope, confidence, verification, collection_status,
 attributes, payload, asset_id, finding_id, exploration_node_id)
SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::uuid,$12,$13,
       $14::uuid,$15,$16,$17,$18,$19,$20::jsonb,$21::jsonb,$22,$23,$24
WHERE EXISTS (
 SELECT 1
 FROM evidence_artifacts ea
 JOIN collection_runs cr ON cr.id=ea.collection_run_id
 WHERE ea.id=$14::uuid
   AND ea.sha256=$15
   AND ea.collection_run_id=$11::uuid
   AND ea.task_id=$12
   AND ea.source_tool_id=$9
   AND cr.task_id=$12
   AND cr.source_tool_id=$9
)
ON CONFLICT (observation_id) DO NOTHING`,
		o.ObservationID, o.IdentityKey, string(o.Type), o.Target, port, o.Protocol, o.Name, o.Value,
		o.SourceToolID, version, o.SourceRunID, o.TaskID, o.ObservedAt,
		evidenceID, o.Evidence[0].Digest, string(o.Scope), string(o.Confidence),
		string(o.Verification), string(o.Collection), jsonString(o.Attributes), string(payload),
		link.AssetID, link.FindingID, link.ExplorationNodeID)
	if err != nil { return false, fmt.Errorf("insert observation claim: %w", err) }
	n, err := res.RowsAffected()
	if err != nil { return false, err }
	if n == 1 { return true, nil }

	var same bool
	err = d.QueryRowContext(ctx, `SELECT payload = $2::jsonb FROM observation_claims WHERE observation_id=$1`, o.ObservationID, string(payload)).Scan(&same)
	if err != nil { return false, fmt.Errorf("check observation replay (evidence registration or run may be missing): %w", err) }
	if !same { return false, fmt.Errorf("observation id collision: %s already exists with a different payload", o.ObservationID) }
	return false, nil
}

func jsonString(v any) string {
	if v == nil { return "{}" }
	b, err := json.Marshal(v)
	if err != nil { return "{}" }
	return string(b)
}
