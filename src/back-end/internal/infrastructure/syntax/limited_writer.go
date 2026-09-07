package syntax

import (
	"bytes"
	"errors"
)

// limitedWriter bounds child-process output before allocation grows unchecked.
type limitedWriter struct {
	buffer bytes.Buffer
	limit  int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if len(p) > w.limit-w.buffer.Len() {
		return 0, errors.New("syntax output limit exceeded")
	}
	return w.buffer.Write(p)
}
