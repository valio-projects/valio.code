// Package fault defines stable boundary errors shared by repository ports and
// application services. Transports map them to sanitized protocol responses.
package fault

import "errors"

var (
	// ErrUnavailable denotes an unconfigured provider or unavailable projection.
	ErrUnavailable = errors.New("capability unavailable")
	// ErrNotFound indicates that the requested identity has no visible record.
	ErrNotFound = errors.New("not found")
	// ErrForbidden rejects a workspace crossing before returning any data.
	ErrForbidden = errors.New("workspace forbidden")
	// ErrInvalid rejects malformed or unsafe input without echoing its content.
	ErrInvalid = errors.New("invalid request")
	// ErrConflict rejects replacement of immutable data or a stale write.
	ErrConflict = errors.New("conflicting immutable identity or concurrent update")
	// ErrScopeTooLarge requires the caller to narrow its explicitly bounded scope.
	ErrScopeTooLarge = errors.New("SCOPE_TOO_LARGE")
)
