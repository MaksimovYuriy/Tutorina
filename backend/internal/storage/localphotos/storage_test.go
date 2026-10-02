package localphotos

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestSaveAndDelete(t *testing.T) {
	root := t.TempDir()
	storage, err := New(root)
	if err != nil {
		t.Fatal(err)
	}

	png := append([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, make([]byte, 32)...)
	publicURL, err := storage.Save(bytes.NewReader(png))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(publicURL, PublicPathPrefix) || !strings.HasSuffix(publicURL, ".png") {
		t.Fatalf("unexpected public URL %q", publicURL)
	}
	files, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("stored files = %d, want 1", len(files))
	}

	if err := storage.Delete(publicURL); err != nil {
		t.Fatal(err)
	}
	files, err = os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("stored files after delete = %d, want 0", len(files))
	}
}

func TestSaveRejectsUnsupportedContent(t *testing.T) {
	storage, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_, err = storage.Save(strings.NewReader("not an image"))
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("error = %v, want ErrUnsupported", err)
	}
}

func TestDeleteRefusesUnmanagedPath(t *testing.T) {
	storage, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := storage.Delete("/etc/passwd"); err == nil {
		t.Fatal("expected unmanaged path error")
	}
}
