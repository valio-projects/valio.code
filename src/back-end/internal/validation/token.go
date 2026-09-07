package validation

import "errors"

// Token checks bounded visible ASCII bearer-token syntax. minimum is the
// receiving service's length contract; no token value appears in an error.
func Token(value string, minimum int) error {
	if minimum < 1 || len(value) < minimum || len(value) > 8192 {
		return errors.New("INVALID_TOKEN: token length is outside the supported bounds")
	}
	for _, c := range value {
		if c <= 32 || c >= 127 {
			return errors.New("INVALID_TOKEN: token must contain visible ASCII without whitespace")
		}
	}
	return nil
}
