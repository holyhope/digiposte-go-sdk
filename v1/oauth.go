package digiposte

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

// TokenSource is a token source that uses a the client to get the oauth token.
type TokenSource struct {
	Client *Client

	GetContext func() context.Context
}

// Token returns a new oauth token.
func (ts *TokenSource) Token() (*oauth2.Token, error) {
	ctx := context.Background()
	if ts.GetContext != nil {
		ctx = ts.GetContext()
	}

	token, err := ts.Client.AccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("access token: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  token.Token,
		Expiry:       token.ExpiresAt,
		TokenType:    "Bearer",
		RefreshToken: "",
	}, nil
}

// TokenSource returns a token source that uses the client to get the oauth token.
func (c *Client) TokenSource() oauth2.TokenSource {
	return oauth2.ReuseTokenSource(nil, &TokenSource{
		Client: c,
	})
}
