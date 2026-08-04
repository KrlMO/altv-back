package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Local struct {
	root    string
	baseURL string
}

func NewLocal(root, baseURL string) *Local {
	return &Local{root: filepath.Clean(root), baseURL: strings.TrimRight(baseURL, "/")}
}

func (l *Local) Save(_ context.Context, folder, filename string, src io.Reader, _ int64, _ string) error {
	dir := filepath.Join(l.root, folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func (l *Local) URL(folder, filename string) string {
	if l.baseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", l.baseURL, strings.Trim(folder, "/"), filename)
}

func (l *Local) Delete(_ context.Context, folder, filename string) error {
	err := os.Remove(filepath.Join(l.root, folder, filename))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
