package storage

import (
	"context"
	"io"
	"strings"
)

type S3 struct {
	baseURL string
}

func NewS3(baseURL string) *S3 {
	return &S3{baseURL: strings.TrimRight(baseURL, "/")}
}

// Save TODO: implement AWS S3 upload (aws-sdk-go-v2) when S3 driver is enabled.
func (s *S3) Save(context.Context, string, string, io.Reader, int64, string) error {
	return ErrNotImplemented
}

func (s *S3) URL(folder, filename string) string {
	if s.baseURL == "" {
		return ""
	}
	return s.baseURL + "/" + strings.Trim(folder, "/") + "/" + filename
}

func (s *S3) Delete(context.Context, string, string) error {
	return ErrNotImplemented
}
