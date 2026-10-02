package localphotos

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	PublicPathPrefix = "/api/v1/media/teacher-photos/"
	MaxFileSize      = 5 << 20
)

var (
	ErrEmpty       = errors.New("photo is empty")
	ErrTooLarge    = errors.New("photo exceeds 5 MB")
	ErrUnsupported = errors.New("photo must be JPEG, PNG or WebP")
)

type Storage struct{ root string }

func New(root string) (*Storage, error) {
	root = filepath.Clean(root)
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create teacher photos directory: %w", err)
	}
	return &Storage{root: root}, nil
}

func (s *Storage) Save(reader io.Reader) (string, error) {
	temporary, err := os.CreateTemp(s.root, ".upload-*")
	if err != nil {
		return "", fmt.Errorf("create temporary photo: %w", err)
	}
	temporaryName := temporary.Name()
	defer func() {
		temporary.Close()
		os.Remove(temporaryName)
	}()

	limited := io.LimitReader(reader, MaxFileSize+1)
	written, err := io.Copy(temporary, limited)
	if err != nil {
		return "", fmt.Errorf("save temporary photo: %w", err)
	}
	if written == 0 {
		return "", ErrEmpty
	}
	if written > MaxFileSize {
		return "", ErrTooLarge
	}
	if _, err := temporary.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("inspect photo: %w", err)
	}
	header := make([]byte, 512)
	read, err := temporary.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("inspect photo: %w", err)
	}
	extension, ok := extensionForType(http.DetectContentType(header[:read]))
	if !ok {
		return "", ErrUnsupported
	}
	identifier := make([]byte, 16)
	if _, err := rand.Read(identifier); err != nil {
		return "", fmt.Errorf("generate photo name: %w", err)
	}
	name := hex.EncodeToString(identifier) + extension
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close temporary photo: %w", err)
	}
	if err := os.Rename(temporaryName, filepath.Join(s.root, name)); err != nil {
		return "", fmt.Errorf("publish photo: %w", err)
	}
	return PublicPathPrefix + name, nil
}

func (s *Storage) Delete(publicURL string) error {
	if publicURL == "" {
		return nil
	}
	if !strings.HasPrefix(publicURL, PublicPathPrefix) {
		return fmt.Errorf("refuse to delete unmanaged photo")
	}
	name := strings.TrimPrefix(publicURL, PublicPathPrefix)
	if name == "" || filepath.Base(name) != name {
		return fmt.Errorf("invalid photo path")
	}
	if err := os.Remove(filepath.Join(s.root, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete photo: %w", err)
	}
	return nil
}

func (s *Storage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" || filepath.Base(name) != name || strings.HasPrefix(name, ".") {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, filepath.Join(s.root, name))
}

func extensionForType(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
