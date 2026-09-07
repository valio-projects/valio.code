package agent

import (
	"errors"
	"github.com/valio-projects/valio.code/internal/validation"
	"regexp"
)

var destinationIdentity = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)

// Validate checks upload configuration without reading source, writing spool or
// contacting a server. Authentication success is verified by the server later.
func (o UploadOptions) Validate() error {
	if _, err := ingestionEndpoint(o.Endpoint); err != nil {
		return err
	}
	if err := validation.Token(o.Token, 1); err != nil {
		return err
	}
	if !destinationIdentity.MatchString(o.WorkspaceID) || !destinationIdentity.MatchString(o.RepositoryID) {
		return errors.New("INVALID_DESTINATION: portable workspace and repository IDs are required")
	}
	return nil
}
