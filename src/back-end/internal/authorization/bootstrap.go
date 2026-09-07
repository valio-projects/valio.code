// Package authorization implements the explicitly local bootstrap principal.
package authorization

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"github.com/valio-projects/valio.code/internal/validation"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Bootstrap authenticates one local bootstrap principal and expiring signed sessions.
type Bootstrap struct {
	tokenHash  [32]byte
	sessionKey [32]byte
	origins    map[string]bool
}

// NewBootstrap validates token and trustedOrigins, then creates an ephemeral session signing key.
func NewBootstrap(token string, trustedOrigins []string) (*Bootstrap, error) {
	if err := validation.Token(token, 32); err != nil {
		return nil, err
	}
	a := &Bootstrap{tokenHash: sha256.Sum256([]byte(token)), origins: map[string]bool{}}
	if _, e := rand.Read(a.sessionKey[:]); e != nil {
		return nil, e
	}
	for _, origin := range trustedOrigins {
		u, e := validation.Endpoint(origin, false)
		if e != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || !(u.Scheme == "https" || u.Scheme == "http") || u.Host == "" {
			return nil, errors.New("invalid trusted origin")
		}
		a.origins[origin] = true
	}
	return a, nil
}

// CheckToken compares token hashes in constant time without retaining the supplied token.
func (a *Bootstrap) CheckToken(token string) bool {
	sum := sha256.Sum256([]byte(token))
	return subtle.ConstantTimeCompare(a.tokenHash[:], sum[:]) == 1
}

// AllowedOrigin checks exact configured origins or the same loopback origin.
// An absent Origin is allowed for bearer CLI calls and safe cookie reads.
func (a *Bootstrap) AllowedOrigin(r *http.Request, required bool) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return !required
	}
	if a.origins[origin] {
		return true
	}
	u, e := url.Parse(origin)
	if e != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || u.Host != r.Host {
		return false
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if u.Scheme != scheme {
		return false
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	return host == "localhost" || (ip != nil && ip.IsLoopback())
}

// Authenticate authenticates r using a bearer token or a valid signed session cookie.
func (a *Bootstrap) Authenticate(r *http.Request) bool {
	if h := r.Header.Get("Authorization"); h != "" {
		return strings.HasPrefix(h, "Bearer ") && a.CheckToken(strings.TrimPrefix(h, "Bearer "))
	}
	c, e := r.Cookie("valio_session")
	if e != nil {
		return false
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 3 {
		return false
	}
	expiry, e := strconv.ParseInt(parts[1], 10, 64)
	if e != nil || time.Now().Unix() > expiry {
		return false
	}
	sig, e := base64.RawURLEncoding.DecodeString(parts[2])
	if e != nil {
		return false
	}
	mac := hmac.New(sha256.New, a.sessionKey[:])
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return false
	}
	unsafe := r.Method != "GET" && r.Method != "HEAD"
	return a.AllowedOrigin(r, unsafe)
}

// SetSession writes an expiring HttpOnly session to w using the transport security of r.
func (a *Bootstrap) SetSession(w http.ResponseWriter, r *http.Request) error {
	nonce := make([]byte, 32)
	if _, e := rand.Read(nonce); e != nil {
		return e
	}
	expiry := time.Now().Add(8 * time.Hour)
	value := base64.RawURLEncoding.EncodeToString(nonce) + "." + strconv.FormatInt(expiry.Unix(), 10)
	mac := hmac.New(sha256.New, a.sessionKey[:])
	mac.Write([]byte(value))
	value += "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	http.SetCookie(w, &http.Cookie{Name: "valio_session", Value: value, Path: "/api/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, Expires: expiry, MaxAge: 8 * 60 * 60})
	return nil
}

// ClearSession expires the browser session cookie on w using the security attributes of r.
func (a *Bootstrap) ClearSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "valio_session", Value: "", Path: "/api/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, MaxAge: -1})
}
