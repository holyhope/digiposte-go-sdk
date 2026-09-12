package partner

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// ClientCredentialsConfig configures the Partner API's client_credentials
// grant (server-to-server, no end-user interaction).
type ClientCredentialsConfig struct {
	// ClientID is the partner application's client ID.
	ClientID string

	// ClientSecret is the partner application's client secret.
	ClientSecret string

	// Endpoint is the Partner API's OAuth2 endpoint. Only TokenURL and
	// AuthStyle are used by this grant.
	Endpoint Endpoint

	// OkapiKey is the value sent as the X-Okapi-Key header on every
	// token request. Required.
	OkapiKey string

	// Scopes optionally specifies requested permissions.
	Scopes []string

	// Timeout bounds every token request made by the returned
	// oauth2.TokenSource. Defaults to 30s when zero.
	Timeout time.Duration
}

// NewClientCredentialsSource returns an [oauth2.TokenSource] that obtains a
// Partner API access token using the client_credentials grant.
//
// ctx is used both to build the returned token source and as the base
// context for every subsequent Token() call it makes; cancelling ctx aborts
// an in-flight request. Every request the returned source makes carries the
// X-Okapi-Key header and is bounded by cfg.Timeout.
func NewClientCredentialsSource( //nolint:ireturn
	ctx context.Context,
	cfg ClientCredentialsConfig,
) (oauth2.TokenSource, error) {
	if cfg.OkapiKey == "" {
		return nil, ErrMissingOkapiKey
	}

	ccCfg := &clientcredentials.Config{
		ClientID:       cfg.ClientID,
		ClientSecret:   cfg.ClientSecret,
		TokenURL:       cfg.Endpoint.TokenURL,
		Scopes:         cfg.Scopes,
		EndpointParams: nil,
		AuthStyle:      cfg.Endpoint.resolvedAuthStyle(),
	}

	tokenCtx := withOkapiHTTPClient(ctx, cfg.OkapiKey, cfg.Timeout)

	source := ccCfg.TokenSource(tokenCtx)

	return &wrappedTokenSource{source: source}, nil
}

// wrappedTokenSource wraps an oauth2.TokenSource to translate errors
// consistently with the rest of this package.
type wrappedTokenSource struct {
	source oauth2.TokenSource
}

var _ oauth2.TokenSource = (*wrappedTokenSource)(nil)

func (s *wrappedTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.source.Token()
	if err != nil {
		return nil, fmt.Errorf("partner: client credentials token: %w", err)
	}

	return token, nil
}
