package main

import (
	"encoding/json"
	"errors"
	"github.com/valio-projects/valio.code/internal/agent"
	"os"
	"path/filepath"
)

func writeOutput(path string, s agent.Snapshot) error {
	payload, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return errors.New("cannot encode snapshot")
	}
	dir := filepath.Dir(path)
	f, e := os.CreateTemp(dir, ".valio-output-*")
	if e != nil {
		return errors.New("cannot create output")
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, e = f.Write(append(payload, '\n')); e != nil {
		f.Close()
		return errors.New("cannot write output")
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return errors.New("cannot sync output")
	}
	if e = f.Close(); e != nil {
		return errors.New("cannot close output")
	}
	if e = os.Rename(temp, path); e != nil {
		return errors.New("cannot finalize output")
	}
	return nil
}
