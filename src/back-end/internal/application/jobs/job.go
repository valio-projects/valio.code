package jobs

import (
	domainjobs "github.com/valio-projects/valio.code/internal/domain/jobs"
	"github.com/valio-projects/valio.code/internal/domain/repositories"
)

// Job aliases the domain record used by the repository contract.
type Job = domainjobs.Job

// ErrLeaseLost denotes an expired, cancelled or superseded fencing token.
var ErrLeaseLost = domainjobs.ErrLeaseLost

// JobRepository is the domain-owned leased job port.
type JobRepository = repositories.JobRepository
