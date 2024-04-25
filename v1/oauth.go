package digiposte

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// TokenSource is a token source that uses a the client to get the oauth token.
type TokenSource struct {
	*clientHelper

	DocumentURL string

	GetContext func() context.Context
}

func NewTokenSource(c *http.Client, documentURL string, getContext func() context.Context) *TokenSource {
	return &TokenSource{
		clientHelper: &clientHelper{client: c},
		DocumentURL:  documentURL,
		GetContext:   getContext,
	}
}

// Token returns a new oauth token.
func (ts *TokenSource) Token() (*oauth2.Token, error) {
	ctx := context.Background()
	if ts.GetContext != nil {
		ctx = ts.GetContext()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.DocumentURL+"/rest/security/token", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	token := new(AccessToken)

	if err := ts.call(req, token); err != nil {
		return nil, fmt.Errorf("call: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  token.Token,
		Expiry:       token.ExpiresAt,
		TokenType:    "Bearer",
		RefreshToken: "",
	}, nil
}
