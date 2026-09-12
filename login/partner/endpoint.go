package partner

import "golang.org/x/oauth2"

// Endpoint describes a Partner API OAuth2 provider's authorization and
// token endpoint URLs, mirroring [oauth2.Endpoint].
//
// The exact URLs are not hardcoded by this package: sandbox and production
// endpoints (and their exact paths) are specific to each partner's own
// OKAPI/Swagger contract, so callers must supply them.
type Endpoint struct {
	// AuthURL is the Partner API's authorization endpoint URL. Only used by
	// the authorization_code grant (see AuthorizationCodeConfig).
	AuthURL string

	// TokenURL is the Partner API's token endpoint URL. Used by both
	// grants.
	TokenURL string

	// AuthStyle optionally specifies how the endpoint wants the client ID
	// and client secret sent when requesting a token.
	//
	// The zero value, oauth2.AuthStyleAutoDetect, is treated as
	// oauth2.AuthStyleInHeader (HTTP Basic) by this package instead of
	// being passed through as auto-detect: the Partner API's documented
	// client_credentials flow is HTTP Basic, and leaving auto-detect in
	// place would let golang.org/x/oauth2 silently retry a failed request
	// with credentials moved into the form body instead of surfacing a
	// single error. Set this explicitly (e.g. oauth2.AuthStyleInParams) if
	// a specific OKAPI space genuinely requires a different style.
	AuthStyle oauth2.AuthStyle
}

// resolvedAuthStyle returns e.AuthStyle, defaulting AuthStyleAutoDetect to
// AuthStyleInHeader.
func (e Endpoint) resolvedAuthStyle() oauth2.AuthStyle {
	if e.AuthStyle == oauth2.AuthStyleAutoDetect {
		return oauth2.AuthStyleInHeader
	}

	return e.AuthStyle
}

// toOAuth2Endpoint converts e into an [oauth2.Endpoint], applying the same
// AuthStyle default as resolvedAuthStyle.
func (e Endpoint) toOAuth2Endpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:       e.AuthURL,
		DeviceAuthURL: "",
		TokenURL:      e.TokenURL,
		AuthStyle:     e.resolvedAuthStyle(),
	}
}
