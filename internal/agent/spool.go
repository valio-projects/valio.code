package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var snapshotID = regexp.MustCompile(`^[0-9a-f]{64}$`)

// MaxSpoolPayloadBytes bounds one on-disk snapshot before JSON decoding. It is
// a transport-resume safety limit, not a claim that a repository corpus cannot
// be larger when captured or uploaded through another protocol.
const MaxSpoolPayloadBytes int64 = 64 * 1024 * 1024

// Save writes one complete sanitized payload using an atomic rename. Repeated
// writes of the same immutable snapshot are idempotent; no raw source spool exists.
func Save(dir string, s Snapshot) error {
	if err := ValidateSnapshot(s); err != nil {
		return err
	}
	if e := os.MkdirAll(dir, 0700); e != nil {
		return errors.New("cannot create spool")
	}
	dst := filepath.Join(dir, s.ID+".json")
	payload, e := json.Marshal(s)
	if e != nil {
		return errors.New("cannot encode snapshot")
	}
	if existing, e := os.ReadFile(dst); e == nil {
		if string(existing) == string(payload) {
			return nil
		}
		return errors.New("spool identity collision")
	}
	f, e := os.CreateTemp(dir, ".pending-*")
	if e != nil {
		return errors.New("cannot create spool payload")
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, e = f.Write(payload); e != nil {
		f.Close()
		return errors.New("cannot write spool payload")
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return errors.New("cannot sync spool payload")
	}
	if e = f.Close(); e != nil {
		return errors.New("cannot close spool payload")
	}
	if e = os.Rename(tmp, dst); e != nil {
		return errors.New("cannot finalize spool payload")
	}
	return nil
}
func validSnapshot(s Snapshot) bool {
	if !snapshotID.MatchString(s.ID) {
		return false
	}
	id := s.ID
	s.ID = ""
	b, e := json.Marshal(s)
	if e != nil {
		return false
	}
	sum := sha256.Sum256(b)
	return id == hex.EncodeToString(sum[:])
}

// Pending validates complete spool payloads and returns deterministic identities.
// Network upload requires an explicit server endpoint through Upload or the CLI.
func Pending(dir string) ([]string, error) {
	entries, e := os.ReadDir(dir)
	if os.IsNotExist(e) {
		return []string{}, nil
	}
	if e != nil {
		return nil, errors.New("cannot read spool")
	}
	ids := []string{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := entry.Name()[:len(entry.Name())-5]
		if !snapshotID.MatchString(id) {
			continue
		}
		s, e := Load(dir, id)
		if e != nil {
			return nil, e
		}
		ids = append(ids, s.ID)
	}
	sort.Strings(ids)
	return ids, nil
}
func Load(dir, id string) (Snapshot, error) {
	if !snapshotID.MatchString(id) {
		return Snapshot{}, errors.New("invalid snapshot ID")
	}
	path := filepath.Join(dir, id+".json")
	info, e := os.Lstat(path)
	if e != nil || !info.Mode().IsRegular() || info.Size() > MaxSpoolPayloadBytes {
		return Snapshot{}, errors.New("invalid spool payload")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return Snapshot{}, errors.New("cannot read spool payload")
	}
	var s Snapshot
	if json.Unmarshal(b, &s) != nil || s.ID != id || ValidateSnapshot(s) != nil {
		return Snapshot{}, errors.New("corrupt spool payload")
	}
	return s, nil
}
