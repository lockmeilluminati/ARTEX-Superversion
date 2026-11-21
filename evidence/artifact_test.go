package evidence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCollectionArtifactRejectsUnsafePaths(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(root, "escape")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	for _, name := range []string{"../secret.txt", filepath.Join(outside, "secret.txt"), "escape"} {
		if f, _, err := openCollectionArtifact(root, name); err == nil {
			f.Close()
			t.Errorf("expected unsafe path %q to be rejected", name)
		}
	}
}

func TestOpenCollectionArtifactRequiresBoundedRegularFile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "directory")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if f, _, err := openCollectionArtifact(root, "directory"); err == nil {
		f.Close()
		t.Fatal("expected directory to be rejected")
	}

	large := filepath.Join(root, "large.bin")
	file, err := os.Create(large)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maxCollectionArtifactBytes + 1); err != nil {
		file.Close()
		t.Fatal(err)
	}
	file.Close()
	if f, _, err := openCollectionArtifact(root, "large.bin"); err == nil {
		f.Close()
		t.Fatal("expected oversized artifact to be rejected")
	}
}

func TestOpenCollectionArtifactAcceptsFileInsideRunDirectory(t *testing.T) {
	root := t.TempDir()
	want := []byte("scanner output")
	if err := os.WriteFile(filepath.Join(root, "scan.jsonl"), want, 0o600); err != nil {
		t.Fatal(err)
	}
	f, info, err := openCollectionArtifact(root, "scan.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if info.Size() != int64(len(want)) {
		t.Fatalf("size = %d, want %d", info.Size(), len(want))
	}
}
