package agent

// SnapshotWriter persists sanitized immutable snapshots. The default DiskSpool
// validates again before an atomic local write; alternate writers are injectable.
type SnapshotWriter interface {
	// Write persists a validated snapshot or returns a safe storage error.
	Write(Snapshot) error
}

// DiskSpool persists snapshots in Dir as atomic JSON files.
type DiskSpool struct{ Dir string }

// Write validates snapshot before atomically persisting it in Dir.
func (s DiskSpool) Write(snapshot Snapshot) error { return Save(s.Dir, snapshot) }

// Read loads a validated snapshot by ID from Dir.
func (s DiskSpool) Read(id string) (Snapshot, error) { return Load(s.Dir, id) }

// Pending lists validated snapshot IDs currently stored in Dir.
func (s DiskSpool) Pending() ([]string, error) { return Pending(s.Dir) }
