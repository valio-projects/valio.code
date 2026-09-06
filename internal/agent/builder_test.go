package agent

import (
	"context"
	"testing"
)

type testReader struct{ Reads []string }

func (r *testReader) Read(path string, limit int64) ([]byte, error) {
	r.Reads = append(r.Reads, path)
	return []byte("password = do-not-store-this"), nil
}

type testWriter struct{ Snapshots []Snapshot }

func (w *testWriter) Write(s Snapshot) error {
	w.Snapshots = append(w.Snapshots, s)
	return ValidateSnapshot(s)
}
func TestBuilderReaderWriterSeamsPreservePolicy(t *testing.T) {
	r := repo(t)
	write(t, r, "safe.go", "package main")
	write(t, r, "credentials.txt", "excluded-before-read")
	reader := &testReader{}
	writer := &testWriter{}
	s, e := (&CaptureBuilder{Options: Options{Root: r}, Reader: reader, Writer: writer}).Build(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	if len(s.Files) != 0 || len(reader.Reads) != 1 || reader.Reads[0] != "safe.go" || len(writer.Snapshots) != 1 {
		t.Fatal("injected readers bypassed policy or writer was not used")
	}
}
