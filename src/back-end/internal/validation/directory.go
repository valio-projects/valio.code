package validation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Directory resolves a local existing directory, including symlinks. It checks
// only the machine executing it; server catalog paths must not use this check.
func Directory(raw string) (string, error) {
	if raw == "" || strings.ContainsAny(raw, "\x00\r\n") {
		return "", errors.New("INVALID_PATH: directory path is empty or malformed")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", errors.New("INVALID_PATH: directory path cannot be resolved")
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errors.New("PATH_NOT_FOUND: directory does not exist")
		}
		return "", errors.New("PATH_UNAVAILABLE: directory cannot be resolved")
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", errors.New("PATH_UNAVAILABLE: directory cannot be inspected")
	}
	if !info.IsDir() {
		return "", errors.New("NOT_A_DIRECTORY: expected an existing directory")
	}
	return resolved, nil
}
