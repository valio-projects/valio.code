package embeddings

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Hash produces a stable SHA-256 identity without retaining supplied text.
func Hash(parts ...string) string { return hash(strings.Join(parts, "\x00")) }

// FingerprintProfile derives a deterministic vector-space identity from profile fields.
// It intentionally excludes Profile.Fingerprint, so reconstructing or validating a profile
// never makes the result depend on an already stored fingerprint.
func FingerprintProfile(p Profile) string {
	return Hash(p.ID, string(p.Kind), fmt.Sprint(p.SchemaVersion), p.Provider, p.Model, p.Revision, p.Distance, fmt.Sprint(p.Dimension))
}

func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func fingerprint(value string) bool {
	_, err := hex.DecodeString(value)
	return err == nil && len(value) == sha256.Size*2
}
