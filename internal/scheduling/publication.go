package scheduling

import "context"

// Publish makes a generation visible only while the same worker still owns a
// live lease. Generation contents must already have been validated and stored.
func (q Queue) Publish(ctx context.Context, j Job, projection, generation string) error {
	_, err := q.DB.Query(ctx, `BEGIN TRANSACTION;
 LET $job=(SELECT * FROM ONLY type::record('job',$key));
 IF $job.status!='running' OR $job.owner!=$owner OR $job.fence!=$fence OR $job.lease_until<=time::unix(time::now()) { THROW 'LEASE_LOST'; };
 UPDATE type::record('job',$key) SET status='completed',lease_until=0;
 UPSERT type::record('projection_head',$projection) SET key=$projection,generation=$generation;
 COMMIT TRANSACTION;`, map[string]any{"key": j.ID, "owner": j.Owner, "fence": j.Fence, "projection": projection, "generation": generation})
	return err
}
