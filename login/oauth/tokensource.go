package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login"
)

// TokenSource is a token source that uses a login method to get the oauth token.
// It logins to Digiposte on every .Token() call. To avoid this, wrap it with oauth2.ReuseTokenSource.
type TokenSource struct {
	// LoginMethod is the login method to use to get the token.
	LoginMethod login.Method

	// Credentials are the credentials to use to get the token.
	Credentials *login.Credentials

	// Listener is the listener to call when the token is updated.
	// It is called with the new token and the cookies.
	// If the listener is nil, it is not called.
	Listener func(token *oauth2.Token, cookies []*http.Cookie)
}

// Token returns a new token.
// It waits for the listener to be called before returning the token.
func (ts *TokenSource) Token() (*oauth2.Token, error) {
	token, cookies, err := ts.LoginMethod.Login(context.Background(), ts.Credentials)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	if ts.Listener != nil {
		ts.Listener(token, cookies)
	}

	return token, nil
}

type CombinedTokenSources []oauth2.TokenSource

func (ts CombinedTokenSources) Token() (*oauth2.Token, error) {
	if len(ts) == 0 {
		return nil, ErrNoTokenSources
	}

	var errs []error

	for i, t := range ts {
		token, err := t.Token()
		if err != nil {
			errs = append(errs, fmt.Errorf("source %d: %w", i+1, err))
		}

		if token.Valid() {
			return token, nil
		}
	}

	return nil, errors.Join(errs...)
}

var ErrNoTokenSources = errors.New("no token sources")
