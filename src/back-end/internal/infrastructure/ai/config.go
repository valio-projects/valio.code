package ai

import (
	"net/http"
	"strings"
	"time"

	"github.com/valio-projects/valio.code/internal/validation"
)

// Config is the fixed provider endpoint and optional bearer credential.
type Config struct {
	Endpoint          string
	Token             string
	Timeout           time.Duration
	AllowInsecureHTTP bool
}

func (c Config) client() (*http.Client, string, error) {
	u, err := validation.Endpoint(c.Endpoint, !c.AllowInsecureHTTP)
	if err != nil {
		return nil, "", err
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errRedirect
		},
	}, strings.TrimRight(u.String(), "/"), nil
}
