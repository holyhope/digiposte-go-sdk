package persistent

import (
	"context"
	"errors"
	"fmt"
	"log"
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

// SaveError wraps an error returned by Store.Save. It is only ever passed to
// an Option's save-error handler (see WithSaveErrorHandler); it is never
// returned from Login, so a Store.Save failure never turns an otherwise
// successful login into a failed one.
type SaveError struct {
	Err error
}

func (e *SaveError) Error() string {
	return fmt.Sprintf("save session: %v", e.Err)
}

func (e *SaveError) Unwrap() error {
	return e.Err
}

// Option customizes the login.Method returned by New.
type Option func(*method)

// WithSaveErrorHandler overrides how a Store.Save error is reported after a
// login has already succeeded. The default handler logs the error via
// log.Default(). The handler is never used to fail Login: per the package
// contract, a save failure is reported separately and does not affect the
// token and cookies Login returns.
func WithSaveErrorHandler(handler func(error)) Option {
	return func(m *method) {
		m.onSaveError = handler
	}
}

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
//     error is reported through the configured save-error handler (see
//     WithSaveErrorHandler) and does not cause Login to return an error.
func New(store Store, newMethod NewMethod, opts ...Option) login.Method { //nolint:ireturn
	resumeMethod := &method{
		store:     store,
		newMethod: newMethod,
		onSaveError: func(err error) {
			log.Printf("digiposte-go-sdk/login/persistent: %v", err)
		},
	}

	for _, opt := range opts {
		opt(resumeMethod)
	}

	return resumeMethod
}

type method struct {
	store       Store
	newMethod   NewMethod
	onSaveError func(error)
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

	saveErr := m.store.Save(ctx, &Session{Token: token, Cookies: cookies})
	if saveErr != nil {
		m.onSaveError(&SaveError{Err: saveErr})
	}

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
