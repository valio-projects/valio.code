package fault

import domainfault "github.com/valio-projects/valio.code/internal/domain/fault"

var (
	// ErrNotFound preserves the corresponding domain boundary error identity.
	ErrNotFound = domainfault.ErrNotFound
	// ErrForbidden preserves the corresponding domain boundary error identity.
	ErrForbidden = domainfault.ErrForbidden
	// ErrInvalid preserves the corresponding domain boundary error identity.
	ErrInvalid = domainfault.ErrInvalid
	// ErrConflict preserves the corresponding domain boundary error identity.
	ErrConflict = domainfault.ErrConflict
	// ErrScopeTooLarge preserves the corresponding domain boundary error identity.
	ErrScopeTooLarge = domainfault.ErrScopeTooLarge
)
