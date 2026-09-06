package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNativeFileChangeSignal(t *testing.T) {
	root := t.TempDir()
	source, err := NewFSNotifySource(root)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	if err = os.WriteFile(filepath.Join(root, "code.go"), []byte("package sample"), 0600); err != nil {
		t.Fatal(err)
	}
	select {
	case <-source.Events():
	case <-time.After(3 * time.Second):
		t.Fatal("native file notification missing")
	}
}
