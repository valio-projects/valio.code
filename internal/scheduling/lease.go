package scheduling

import (
	"context"
	"encoding/json"
)

func (q Queue) Heartbeat(ctx context.Context, j Job) error {
	return q.mutate(ctx, j, `UPDATE type::record('job',$key) SET lease_until=time::unix(time::now())+60
 RETURN AFTER;`, nil)
}
func (q Queue) Complete(ctx context.Context, j Job) error {
	return q.mutate(ctx, j, `UPDATE type::record('job',$key) SET status='completed',lease_until=0
 RETURN AFTER;`, nil)
}
func (q Queue) Fail(ctx context.Context, j Job, code string, temporary bool) error {
	status := "failed"
	if temporary && j.Attempts <= 5 {
		status = "queued"
	}
	delay := int64(1 << min(j.Attempts, 8))
	return q.mutate(ctx, j, `UPDATE type::record('job',$key) SET status=$status,error_code=$code,
 available_at=time::unix(time::now())+$delay,lease_until=0
 RETURN AFTER;`, map[string]any{"status": status, "code": code, "delay": delay})
}
func (q Queue) Cancel(ctx context.Context, key string) error {
	_, err := q.DB.Query(ctx, `BEGIN TRANSACTION;
 LET $job=(SELECT * FROM ONLY type::record('job',$key));
 IF $job.status IN ['queued','running'] { UPDATE type::record('job',$key) SET status='cancelled',fence=fence+1,lease_until=0; };
 COMMIT TRANSACTION;`, map[string]any{"key": key})
	return err
}
func (q Queue) mutate(ctx context.Context, j Job, sql string, extra map[string]any) error {
	vars := map[string]any{"key": j.ID, "owner": j.Owner, "fence": j.Fence}
	for k, v := range extra {
		vars[k] = v
	}
	// 3.2.x UPDATE WHERE can observe fields modified by SET. Check the saved
	// record in an explicit transaction before mutation instead.
	rows, err := q.DB.Query(ctx, `BEGIN TRANSACTION;
 LET $job=(SELECT * FROM ONLY type::record('job',$key));
 LET $changed=IF $job.status='running' AND $job.owner=$owner AND $job.fence=$fence AND $job.lease_until>time::unix(time::now()) { `+sql+` } ELSE { [] };
 RETURN $changed;
 COMMIT TRANSACTION;`, vars)
	if err != nil {
		return err
	}
	for _, row := range rows {
		var jobs []Job
		if json.Unmarshal(row.Result, &jobs) == nil && len(jobs) == 1 && jobs[0].ID == j.ID {
			return nil
		}
	}
	return ErrLeaseLost
}
