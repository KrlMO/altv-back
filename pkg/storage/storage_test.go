package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalSaveURLDelete(t *testing.T) {
	root := t.TempDir()
	s := NewLocal(root, "http://localhost:8080/media")

	ctx := context.Background()
	content := strings.NewReader("fake-image-bytes")

	if err := s.Save(ctx, "teams/logo", "abc.png", content, 0, "image/png"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, "teams", "logo", "abc.png"))
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if string(got) != "fake-image-bytes" {
		t.Fatalf("unexpected content: %s", got)
	}

	if url := s.URL("teams/logo", "abc.png"); url != "http://localhost:8080/media/teams/logo/abc.png" {
		t.Fatalf("unexpected URL: %s", url)
	}

	if err := s.Delete(ctx, "teams/logo", "abc.png"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "teams", "logo", "abc.png")); !os.IsNotExist(err) {
		t.Fatalf("file not deleted")
	}
}

func TestRandomFilename(t *testing.T) {
	a := RandomFilename(".png")
	b := RandomFilename(".png")
	if a == b {
		t.Fatalf("expected unique filenames, got %q twice", a)
	}
	if !strings.HasSuffix(a, ".png") || strings.Contains(a, "/") {
		t.Fatalf("unexpected filename: %s", a)
	}
}
