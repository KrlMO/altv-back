package storage

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

func RandomFilename(ext string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b) + strings.ToLower(ext)
}
