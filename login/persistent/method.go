package persistent

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login"
)

// NewMethod builds the login.Method to use when no valid stored session is
// available. seedCookies contains the cookies from the most recently loaded
// session, if any (nil if none was stored, or if loading failed); a factory
// can seed them into its returned login.Method - for example
// chrome.WithCookies(seedCookies) - so that a still-authenticated site
// session is recognized without re-entering credentials.
type NewMethod func(seedCookies []*http.Cookie) (login.Method, error)

// errNilStore is returned by Login when New was called with a nil Store.
// New itself does not validate this eagerly, to keep its signature exactly
// New(store, newMethod) login.Method with no error return: a nil Store is a
// caller bug, not a runtime condition to design around, and is reported the
// first time it would actually be used.
var errNilStore = errors.New("persistent: store must not be nil")

// New returns a login.Method that resumes a previously stored session via
// store when it is still valid, instead of calling newMethod. store must not
// be nil; a caller that wants a plain, never-persisted login should simply
// not use this package.
//
// On every call to Login:
//   - store.Load is called. If it returns a Session whose Token is still
//     valid (per oauth2.Token.Valid), that token and its cookies are
//     returned directly and newMethod is never called.
//   - Otherwise - no stored session, an expired/invalid token, or a Load
//     error - newMethod is called with the loaded cookies (nil if there
//     were none), and Login delegates to the resulting login.Method.
//   - After a successful login, whether resumed or obtained from newMethod,
//     the resulting token and cookies are saved via store.Save. A Save
//     error is discarded: Store is caller-supplied, so the caller's own
//     Save implementation already has the error and is responsible for
//     handling it (logging, retrying, alerting); Login never returns it.
func New(store Store, newMethod NewMethod) login.Method { //nolint:ireturn
	return &method{
		store:     store,
		newMethod: newMethod,
	}
}

type method struct {
	store     Store
	newMethod NewMethod
}

var _ login.Method = (*method)(nil)

func (m *method) Login(ctx context.Context, creds *login.Credentials) (*oauth2.Token, []*http.Cookie, error) {
	if m.store == nil {
		return nil, nil, errNilStore
	}

	token, cookies, err := m.resolve(ctx, creds)
	if err != nil {
		return nil, nil, err
	}

	_ = m.store.Save(ctx, &Session{Token: token, Cookies: cookies})

	return token, cookies, nil
}

// resolve returns a valid token and cookies, either resumed from the store
// or obtained by delegating to newMethod. It does not persist the result;
// that is Login's responsibility.
func (m *method) resolve(
	ctx context.Context,
	creds *login.Credentials,
) (*oauth2.Token, []*http.Cookie, error) {
	var seedCookies []*http.Cookie

	// A Load error - whether the documented ErrSessionNotFound or anything
	// else the Store implementation returns - is treated the same way: fall
	// back to an interactive login with no seed cookies, rather than failing
	// the overall login attempt.
	session, err := m.store.Load(ctx)
	if err == nil && session != nil {
		seedCookies = session.Cookies

		if session.Token != nil && session.Token.Valid() {
			return session.Token, session.Cookies, nil
		}
	}

	innerMethod, err := m.newMethod(seedCookies)
	if err != nil {
		return nil, nil, fmt.Errorf("new login method: %w", err)
	}

	token, cookies, err := innerMethod.Login(ctx, creds)
	if err != nil {
		return nil, nil, fmt.Errorf("login: %w", err)
	}

	return token, cookies, nil
}
