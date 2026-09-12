// Package partner authenticates against Digiposte's official Partner API v3
// (https://developer.laposte.fr/catalog-apis/digiposte@3), fronted by La
// Poste's OKAPI gateway.
//
// This package is entirely independent of the [github.com/holyhope/digiposte-go-sdk/login]
// package and its self-vault login methods (login/chrome, login/persistent):
// it authenticates as a registered Digiposte Partenaire (an organization),
// not as a specific end user browsing their own vault, uses real OAuth2
// instead of a browser-scraped session, and every request additionally
// requires an X-Okapi-Key header.
//
// Two independent ways to obtain a token are provided, matching the two
// grants documented by the Partner API:
//
//   - [NewClientCredentialsSource] for the client_credentials grant
//     (server-to-server, no end-user interaction).
//   - [NewAuthorizationCodeFlow] for the authorization_code grant with PKCE
//     (requires a specific Digiposte end user's consent).
//
// This package only produces an *oauth2.Token usable as a Bearer token
// against the Partner API - it does not implement any Partner API resource
// endpoint (memberships/adhésions, document deposit/retrieval).
package partner
