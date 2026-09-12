package partner

import (
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// OkapiKeyHeader is the HTTP header the Partner API (via La Poste's OKAPI
// gateway) requires on every request. This is a header name, not a secret
// value.
const OkapiKeyHeader = "X-Okapi-Key" //nolint:gosec

// ErrMissingOkapiKey is returned when a grant is configured without an
// Okapi key, before any request is made to the Partner API.
var ErrMissingOkapiKey = errors.New("partner: missing Okapi key")

// defaultTimeout is applied to the internal *http.Client used for token
// requests when a Config's Timeout field is left at zero.
const defaultTimeout = 30 * time.Second

// okapiTransport injects the X-Okapi-Key header into every outgoing
// request, without mutating the caller-supplied base RoundTripper or the
// original request.
type okapiTransport struct {
	key  string
	base http.RoundTripper
}

var _ http.RoundTripper = (*okapiTransport)(nil)

func (t *okapiTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set(OkapiKeyHeader, t.key)

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	resp, err := base.RoundTrip(clone)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return resp, nil
}

// okapiHTTPClient builds an *http.Client that injects the X-Okapi-Key
// header on every request it makes, bounded by timeout (defaultTimeout
// when zero).
func okapiHTTPClient(okapiKey string, timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	return &http.Client{
		Transport:     &okapiTransport{key: okapiKey, base: nil},
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       timeout,
	}
}

// withOkapiHTTPClient installs an Okapi-key-injecting, timeout-bounded
// *http.Client into ctx via the golang.org/x/oauth2 convention
// (oauth2.HTTPClient), so any token request made with the returned context
// through the oauth2/clientcredentials libraries carries the header.
func withOkapiHTTPClient(ctx context.Context, okapiKey string, timeout time.Duration) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, okapiHTTPClient(okapiKey, timeout))
}
