package agent

// SnapshotWriter persists sanitized immutable snapshots. The default DiskSpool
// validates again before an atomic local write; alternate writers are injectable.
type SnapshotWriter interface{ Write(Snapshot) error }
type DiskSpool struct{ Dir string }

func (s DiskSpool) Write(snapshot Snapshot) error    { return Save(s.Dir, snapshot) }
func (s DiskSpool) Read(id string) (Snapshot, error) { return Load(s.Dir, id) }
func (s DiskSpool) Pending() ([]string, error)       { return Pending(s.Dir) }
