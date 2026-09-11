// Package persistent lets an SDK caller reuse a previously obtained
// Digiposte login session (an OAuth token and its cookies) across process
// runs, instead of always driving an interactive, browser-based login.
package persistent

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/oauth2"
)

// Session is a previously obtained login session: an OAuth token and the
// cookies associated with it.
type Session struct {
	Token   *oauth2.Token
	Cookies []*http.Cookie
}

// ErrSessionNotFound is returned by Store.Load when no session has been
// saved yet. It is not treated as an error by the login.Method returned
// from New: falling back to an interactive login is the expected behavior
// in that case.
var ErrSessionNotFound = errors.New("session not found")

// Store loads and saves a Session across process runs.
//
// A stored Session is equivalent to live account credentials: whoever can
// read it can act as the authenticated user until the token expires or the
// cookies are invalidated. Implementations MUST protect stored sessions at
// least as carefully as a password (for example: restrictive file
// permissions, encryption at rest, or a dedicated secrets manager) and MUST
// NOT log, print, or otherwise expose their contents. This package
// intentionally ships no default Store implementation; callers must supply
// their own.
type Store interface {
	// Load returns the previously saved Session, or a wrapped
	// ErrSessionNotFound (checked with errors.Is) if none has been saved
	// yet. Any other error is treated the same as "no session found" by the
	// login.Method returned from New: it falls back to an interactive
	// login rather than failing outright.
	Load(ctx context.Context) (*Session, error)

	// Save persists session for later retrieval by Load.
	Save(ctx context.Context, session *Session) error
}
