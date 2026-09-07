// Package syntax adapts a pinned Node/WASM syntax helper to snapshot ingestion.
package syntax

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/valio-projects/valio.code/internal/domain"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Processor owns a fixed helper executable, never a command supplied by a query.
type Processor struct {
	node    string
	script  string
	profile string
}

// New validates tool paths and fingerprints the helper source and grammar lock.
func New(node, script string) (*Processor, error) {
	binary, err := exec.LookPath(node)
	if err != nil {
		return nil, errors.New("SYNTAX_RUNTIME_UNAVAILABLE")
	}
	path, err := filepath.Abs(script)
	if err != nil {
		return nil, errors.New("INVALID_SYNTAX_HELPER_PATH")
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("SYNTAX_HELPER_NOT_FOUND")
	}
	hash := sha256.New()
	total := 0
	err = filepath.WalkDir(filepath.Dir(path), func(name string, entry os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if entry.IsDir() {
			if entry.Name() == "node_modules" || entry.Name() == "test" || entry.Name() == "tests" {
				return filepath.SkipDir
			}
			return nil
		}
		if !(strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".mjs") || entry.Name() == "package-lock.json") {
			return nil
		}
		data, e := os.ReadFile(name)
		if e != nil {
			return e
		}
		total += len(data)
		if total > 2<<20 {
			return errors.New("helper source budget exceeded")
		}
		rel, _ := filepath.Rel(filepath.Dir(path), name)
		hash.Write([]byte(filepath.ToSlash(rel)))
		hash.Write([]byte{0})
		hash.Write(data)
		return nil
	})
	if err != nil {
		return nil, errors.New("SYNTAX_HELPER_UNREADABLE")
	}
	processor := &Processor{node: binary, script: path, profile: "wasm-syntax/v1:" + hex.EncodeToString(hash.Sum(nil))}
	// Check the actual runtime and grammar at startup, including native libraries.
	// A present executable path alone does not prove the analyzer can run.
	probe, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := processor.Analyze(probe, "probe.ts", "typescript", "export class Probe {}"); err != nil {
		return nil, errors.New("SYNTAX_STARTUP_PROBE_FAILED")
	}
	return processor, nil
}

// Profile pins the analyzer implementation independently of source versions.
func (p *Processor) Profile() string { return p.profile }

// Analyze parses one source in an isolated bounded process. Stderr is discarded
// so runtime exceptions cannot echo source or local configuration into logs.
func (p *Processor) Analyze(ctx context.Context, path, language, content string) (json.RawMessage, error) {
	if domain.ValidateRelativePath(path, false) != nil || len(content) > 2<<20 {
		return nil, errors.New("INVALID_SYNTAX_INPUT")
	}
	child, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	input, err := json.Marshal(map[string]any{"schema": 1, "path": path, "language": language, "content": content})
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(child, p.node, "--max-old-space-size=256", p.script)
	cmd.Stdin = bytes.NewReader(append(input, '\n'))
	output := &limitedWriter{limit: 8 << 20}
	cmd.Stdout = output
	cmd.Stderr = io.Discard
	if err = cmd.Run(); err != nil {
		if child.Err() != nil {
			return nil, child.Err()
		}
		return nil, errors.New("SYNTAX_ANALYZER_FAILED")
	}
	payload := json.RawMessage(bytes.Clone(output.buffer.Bytes()))
	if err = validateReport(payload, path, language, len(content)); err != nil {
		return nil, err
	}
	return payload, nil
}
