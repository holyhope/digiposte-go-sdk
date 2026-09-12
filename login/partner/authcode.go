package partner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

// ErrStateMismatch is returned by (*AuthorizationCodeFlow).Exchange when the
// state supplied does not match the state used to build the authorization
// URL. No request is made to the Partner API when this error is returned.
var ErrStateMismatch = errors.New("partner: state mismatch")

// AuthorizationCodeConfig configures the Partner API's authorization_code
// grant with PKCE (requires a specific Digiposte end user's consent).
type AuthorizationCodeConfig struct {
	// ClientID is the partner application's client ID.
	ClientID string

	// ClientSecret is the partner application's client secret.
	ClientSecret string

	// Endpoint is the Partner API's OAuth2 endpoint.
	Endpoint Endpoint

	// RedirectURL is the URL Digiposte redirects the end user back to
	// after consent, previously declared to Digiposte for this partner.
	RedirectURL string

	// OkapiKey is the value sent as the X-Okapi-Key header on every
	// token request. Required.
	OkapiKey string

	// Scopes optionally specifies requested permissions.
	Scopes []string

	// Timeout bounds the token request made by Exchange. Defaults to 30s
	// when zero.
	Timeout time.Duration
}

// AuthorizationCodeFlow drives a single authorization_code + PKCE round
// trip: build an authorization URL with [AuthorizationCodeFlow.AuthCodeURL],
// redirect the end user to it, then exchange the code Digiposte returns
// with [AuthorizationCodeFlow.Exchange].
//
// A flow value must survive between the AuthCodeURL and Exchange calls. If
// the process that builds the URL may not be the one that receives the
// callback (e.g. a horizontally-scaled web server), persist State and
// CodeVerifier yourself (for example keyed by State in your own session
// store) and reconstruct an equivalent flow value before calling Exchange.
type AuthorizationCodeFlow struct {
	cfg AuthorizationCodeConfig

	// State is the CSRF state value embedded in the authorization URL.
	// Exchange rejects any state that does not match this value.
	State string

	// CodeVerifier is the PKCE code verifier generated for this flow. Its
	// S256 challenge is embedded in the authorization URL, and it is sent
	// again, verbatim, to the token endpoint by Exchange.
	CodeVerifier string
}

// NewAuthorizationCodeFlow returns a new [AuthorizationCodeFlow], generating
// a fresh State and CodeVerifier.
func NewAuthorizationCodeFlow(cfg AuthorizationCodeConfig) (*AuthorizationCodeFlow, error) {
	if cfg.OkapiKey == "" {
		return nil, ErrMissingOkapiKey
	}

	return &AuthorizationCodeFlow{
		cfg:          cfg,
		State:        generateState(),
		CodeVerifier: oauth2.GenerateVerifier(),
	}, nil
}

// generateState returns a fresh, cryptographically random CSRF state value,
// with the same strength as oauth2.GenerateVerifier.
func generateState() string {
	return oauth2.GenerateVerifier()
}

// AuthCodeURL returns the URL to redirect the end user to, embedding
// f.State and a PKCE code_challenge derived from f.CodeVerifier.
func (f *AuthorizationCodeFlow) AuthCodeURL() string {
	return f.oauth2Config().AuthCodeURL(f.State, oauth2.S256ChallengeOption(f.CodeVerifier))
}

// Exchange exchanges code (returned by Digiposte on the redirect callback)
// for a Partner API access token. state must match f.State - a mismatch is
// treated as [ErrStateMismatch] and no request is made to the Partner API.
//
// ctx bounds this call in addition to f.cfg.Timeout; cancelling it aborts an
// in-flight request.
func (f *AuthorizationCodeFlow) Exchange(ctx context.Context, code, state string) (*oauth2.Token, error) {
	if state != f.State {
		return nil, ErrStateMismatch
	}

	tokenCtx := withOkapiHTTPClient(ctx, f.cfg.OkapiKey, f.cfg.Timeout)

	token, err := f.oauth2Config().Exchange(tokenCtx, code, oauth2.VerifierOption(f.CodeVerifier))
	if err != nil {
		return nil, fmt.Errorf("partner: authorization code exchange: %w", err)
	}

	return token, nil
}

func (f *AuthorizationCodeFlow) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     f.cfg.ClientID,
		ClientSecret: f.cfg.ClientSecret,
		Endpoint:     f.cfg.Endpoint.toOAuth2Endpoint(),
		RedirectURL:  f.cfg.RedirectURL,
		Scopes:       f.cfg.Scopes,
	}
}
