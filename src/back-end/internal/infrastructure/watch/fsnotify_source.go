package watch

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FSNotifySource adapts fsnotify into coalesced signals. It never reads file contents.
type FSNotifySource struct {
	watcher  *fsnotify.Watcher
	events   chan struct{}
	failures chan error
}

// NewFSNotifySource recursively watches root without following symlinks. Common
// generated directories are left to full reconciliation to bound watch handles.
func NewFSNotifySource(root string) (*FSNotifySource, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, errors.New("invalid watch root")
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.New("filesystem notifications unavailable")
	}
	source := &FSNotifySource{watcher: watcher, events: make(chan struct{}, 1), failures: make(chan error, 1)}
	if err = source.addTree(absolute); err != nil {
		watcher.Close()
		return nil, errors.New("watch registration incomplete; use reconciliation")
	}
	go source.pump()
	return source, nil
}
func (s *FSNotifySource) addTree(root string) error {
	count := len(s.watcher.WatchList())
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		name := strings.ToLower(entry.Name())
		if name == ".git" || name == ".valio" || name == ".tools" || name == "node_modules" {
			return filepath.SkipDir
		}
		if count >= 8192 {
			return errors.New("watch handle budget exceeded")
		}
		count++
		return s.watcher.Add(path)
	})
}
func (s *FSNotifySource) pump() {
	defer close(s.events)
	defer close(s.failures)
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) {
				if info, err := os.Lstat(event.Name); err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
					if s.addTree(event.Name) != nil {
						s.report()
					}
				}
			}
			select {
			case s.events <- struct{}{}:
			default:
			}
		case _, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.report()
		}
	}
}
func (s *FSNotifySource) report() {
	select {
	case s.failures <- errors.New("filesystem notification gap; reconciling"):
	default:
	}
}

// Events returns a coalesced hint channel; dropped duplicate hints are intentional.
func (s *FSNotifySource) Events() <-chan struct{} { return s.events }

// Errors reports notification loss without exposing filesystem payloads.
func (s *FSNotifySource) Errors() <-chan error { return s.failures }

// Close releases platform watch handles and terminates the adapter goroutine.
func (s *FSNotifySource) Close() error { return s.watcher.Close() }
