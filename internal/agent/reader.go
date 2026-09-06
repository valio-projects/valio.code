package agent

import (
	"errors"
	"io"
	"os"
)

// SourceReader supplies bounded source bytes; CaptureBuilder applies the same
// mandatory filtering policy to every implementation before any content hashes.
type SourceReader interface {
	// Read returns repository-relative bytes up to limit or a safe error.
	Read(path string, limit int64) ([]byte, error)
}

// FilesystemReader reads through Root so paths cannot escape the worktree.
type FilesystemReader struct{ Root *os.Root }

// Read rejects non-regular or oversize files beneath Root.
func (r FilesystemReader) Read(path string, limit int64) ([]byte, error) {
	f, e := r.Root.Open(path)
	if e != nil {
		return nil, errors.New("read failed")
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() {
		return nil, errors.New("non-regular file")
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) > limit {
		return nil, errors.New("read limit exceeded")
	}
	return b, nil
}
