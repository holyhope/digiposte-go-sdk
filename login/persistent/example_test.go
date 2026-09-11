package persistent_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login"
	"github.com/holyhope/digiposte-go-sdk/login/chrome"
	"github.com/holyhope/digiposte-go-sdk/login/persistent"
)

// exampleStore is a minimal Store backed by a package-level variable, for
// illustration only. persistent ships no default Store implementation - see
// the Store doc comment for why - so a real caller must write one of these
// backed by a file, an OS keychain, a secrets manager, etc., and must
// protect it at least as carefully as a password.
type exampleStore struct {
	session *persistent.Session
}

func (s *exampleStore) Load(_ context.Context) (*persistent.Session, error) {
	if s.session == nil {
		return nil, persistent.ErrSessionNotFound
	}

	return s.session, nil
}

func (s *exampleStore) Save(_ context.Context, session *persistent.Session) error {
	s.session = session

	return nil
}

// ExampleNew shows how to compose the chrome login method (see
// login/oauth's ExampleTokenSource for the same composition without session
// persistence) with a Store so that a still-valid stored session is resumed
// without ever driving an interactive, browser-based login.
//
// This example's store already holds a valid session, so the chrome-backed
// factory below is never actually called - which is the point: once a
// session has been obtained once and saved, later runs skip Chrome
// entirely as long as the stored token remains valid.
func ExampleNew() {
	store := &exampleStore{
		session: &persistent.Session{
			Token: &oauth2.Token{
				AccessToken:  "previously-obtained-token",
				TokenType:    "Bearer",
				RefreshToken: "",
				Expiry:       time.Now().Add(time.Hour),
				ExpiresIn:    0,
			},
			Cookies: []*http.Cookie{
				{
					Name:        "session",
					Value:       "previously-obtained-cookie",
					Path:        "",
					Domain:      "",
					Expires:     time.Time{},
					RawExpires:  "",
					MaxAge:      0,
					Secure:      true,
					HttpOnly:    true,
					SameSite:    http.SameSiteLaxMode,
					Partitioned: false,
					Raw:         "",
					Unparsed:    nil,
					Quoted:      false,
				},
			},
		},
	}

	loginMethod := persistent.New(store, func(seedCookies []*http.Cookie) (login.Method, error) {
		// Only called when no valid session is stored: seed whatever
		// cookies were loaded (possibly none) into a fresh interactive
		// chrome login, so a still-authenticated site session is
		// recognized without re-entering credentials.
		return chrome.New(
			chrome.WithURL(os.Getenv("DIGIPOSTE_URL")),
			chrome.WithCookies(seedCookies),
		)
	})

	token, _, err := loginMethod.Login(context.Background(), &login.Credentials{
		Username:  os.Getenv("DIGIPOSTE_USERNAME"),
		Password:  os.Getenv("DIGIPOSTE_PASSWORD"),
		OTPSecret: os.Getenv("DIGIPOSTE_OTP_SECRET"),
	})
	if err != nil {
		panic(fmt.Errorf("login: %w", err))
	}

	fmt.Printf("Resumed session, token valid: %v\n", token.Valid())

	// Output:
	// Resumed session, token valid: true
}
