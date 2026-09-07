// Package validation provides reusable boundary checks without echoing inputs.
package validation

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Endpoint parses an absolute HTTP(S) base URL. Credentials, queries, fragments,
// invalid ports and ambiguous path traversal are rejected before network I/O.
// localHTTP limits plaintext HTTP to explicitly named loopback addresses.
func Endpoint(raw string, localHTTP bool) (*url.URL, error) {
	invalid := errors.New("INVALID_ENDPOINT: expected an absolute HTTP(S) URL without credentials, query or fragment")
	if len(raw) > 4096 || strings.TrimSpace(raw) != raw || strings.ContainsAny(raw, "\\\r\n\t") {
		return nil, invalid
	}
	u, err := url.Parse(raw)
	if err != nil || u.Opaque != "" || u.User != nil || u.Hostname() == "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(raw, "#") || (u.Scheme != "https" && u.Scheme != "http") {
		return nil, invalid
	}
	host := strings.ToLower(u.Hostname())
	if net.ParseIP(host) == nil {
		if len(host) > 253 {
			return nil, invalid
		}
		for _, label := range strings.Split(host, ".") {
			if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return nil, invalid
			}
			for _, c := range label {
				if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
					return nil, invalid
				}
			}
		}
	}
	if strings.HasSuffix(u.Host, ":") {
		return nil, invalid
	}
	if p := u.Port(); p != "" {
		n, e := strconv.Atoi(p)
		if e != nil || n < 1 || n > 65535 {
			return nil, invalid
		}
	}
	if strings.Contains(u.Path, "\\") {
		return nil, invalid
	}
	for _, segment := range strings.Split(u.Path, "/") {
		if segment == "." || segment == ".." {
			return nil, invalid
		}
	}
	if localHTTP && u.Scheme == "http" && host != "localhost" && host != "127.0.0.1" && host != "::1" {
		return nil, errors.New("INVALID_ENDPOINT: plaintext HTTP is limited to loopback")
	}
	return u, nil
}
