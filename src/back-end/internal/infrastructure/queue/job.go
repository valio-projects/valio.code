package queue

import domainjobs "github.com/valio-projects/valio.code/internal/domain/jobs"

// Job aliases the domain record used by the repository contract.
type Job = domainjobs.Job

// ErrLeaseLost denotes an expired, cancelled or superseded fencing token.
var ErrLeaseLost = domainjobs.ErrLeaseLost
