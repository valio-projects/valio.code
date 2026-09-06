package scheduling

import (
	"encoding/json"
	"errors"
)

var ErrLeaseLost = errors.New("job lease lost or cancelled")

type Job struct {
	ID         string          `json:"key"`
	Kind       string          `json:"kind"`
	Payload    json.RawMessage `json:"payload"`
	Status     string          `json:"status"`
	Priority   int             `json:"priority"`
	Attempts   int             `json:"attempts"`
	Fence      int             `json:"fence"`
	Owner      string          `json:"owner"`
	LeaseUntil int64           `json:"lease_until"`
	ErrorCode  string          `json:"error_code,omitempty"`
}
