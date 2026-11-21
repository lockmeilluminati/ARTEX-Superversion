package evidence

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Autumn-27/artex/db"
)

const maxCollectionArtifactBytes int64 = 256 << 20

// Artifact is the opaque reference returned after an output file has been
// copied into the evidence store and its bytes have been hashed.
type Artifact struct {
	ID         string    `json:"id"`
	SHA256     string    `json:"sha256"`
	ByteLength int64     `json:"byte_length"`
	MediaType  string    `json:"media_type"`
	CapturedAt time.Time `json:"captured_at"`
}

// openCollectionArtifact resolves a regular file beneath root. Symlinks are
// resolved before opening and the resolved path must remain inside the root.
func openCollectionArtifact(root, name string) (*os.File, fs.FileInfo, error) {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(name) == "" {
		return nil, nil, errors.New("artifact root and path are required")
	}
	for _, part := range strings.Split(filepath.ToSlash(name), "/") {
		if part == ".." {
			return nil, nil, errors.New("artifact path must not contain ..")
		}
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve artifact root: %w", err)
	}
	candidate := name
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(realRoot, candidate)
	}
	realPath, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return nil, nil, errors.New("artifact path does not exist beneath its run directory")
	}
	rel, err := filepath.Rel(realRoot, realPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, nil, errors.New("artifact path resolves outside its run directory")
	}
	f, err := os.Open(realPath)
	if err != nil {
		return nil, nil, errors.New("artifact file cannot be opened")
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, nil, errors.New("artifact must be a regular file")
	}
	if info.Size() < 0 || info.Size() > maxCollectionArtifactBytes {
		f.Close()
		return nil, nil, fmt.Errorf("artifact exceeds the %d-byte limit", maxCollectionArtifactBytes)
	}
	return f, info, nil
}

// StoreCollectionArtifact imports a scanner output file into the existing
// content-addressed evidence store. The digest and length are computed from
// the copied bytes; callers cannot supply or override them. The collection run
// must already exist. A failed metadata insert can leave an unreferenced blob,
// which the normal evidence GC reaps after its grace period.
func (s *Store) StoreCollectionArtifact(ctx context.Context, runID string, taskID int64, sourceToolID, runDir, relativePath, mediaType string) (Artifact, error) {
	var out Artifact
	if s == nil || s.DB == nil {
		return out, errors.New("evidence database is not initialized")
	}
	if taskID <= 0 || strings.TrimSpace(runID) == "" || strings.TrimSpace(sourceToolID) == "" {
		return out, errors.New("task, collection run and source tool are required")
	}
	f, info, err := openCollectionArtifact(runDir, relativePath)
	if err != nil {
		return out, err
	}
	defer f.Close()

	hash, err := s.writeBody(f, info.Size(), "")
	if err != nil {
		return out, fmt.Errorf("store collection artifact bytes: %w", err)
	}
	id, err := db.NewCollectionRunID()
	if err != nil {
		return out, err
	}
	now := time.Now().UTC()
	if mediaType == "" {
		mediaType = "application/octet-stream"
	}
	task := taskID
	run := runID
	err = s.DB.RegisterEvidenceArtifactMetadata(ctx, db.EvidenceArtifactMetadata{
		ID: id, SHA256: hash, ByteLength: info.Size(), MediaType: mediaType,
		StorageKey: hash, TaskID: &task, CollectionRunID: &run,
		SourceToolID: sourceToolID, OriginalName: filepath.Base(relativePath), CapturedAt: now,
	})
	if err != nil {
		return out, fmt.Errorf("register collection artifact: %w", err)
	}
	return Artifact{ID: id, SHA256: hash, ByteLength: info.Size(), MediaType: mediaType, CapturedAt: now}, nil
}
