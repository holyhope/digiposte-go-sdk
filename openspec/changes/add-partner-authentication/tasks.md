## 1. Package scaffolding

- [ ] 1.1 Create `login/partner` package with a doc comment explaining its scope (Partner API OAuth2 authentication only, independent of `login`) and verify `GOWORK=off go build ./login/partner/...` succeeds with an empty package
- [ ] 1.2 Add `Endpoint` config type (`AuthURL`, `TokenURL` string fields, `AuthStyle oauth2.AuthStyle`, mirroring `oauth2.Endpoint`) shared by both grants, and verify it round-trips into `oauth2.Config.Endpoint`/`clientcredentials.Config.{TokenURL,AuthStyle}` via a small unit test, including that an unset `AuthStyle` maps to `oauth2.AuthStyleInHeader` rather than `oauth2.AuthStyleAutoDetect`

## 2. Okapi key transport

- [ ] 2.1 Implement an `http.RoundTripper` that injects `X-Okapi-Key: <key>` into every outgoing request without mutating the caller-supplied base transport, and verify with a unit test using `httptest.Server` asserting the header is present
- [ ] 2.2 Implement a helper that builds an `*http.Client` using that transport and installs it via `context.WithValue(ctx, oauth2.HTTPClient, client)`, and verify with a unit test that a token request made through the resulting context reaches the Okapi-key-wrapped transport
- [ ] 2.3 Add a configuration-error path (e.g. `ErrMissingOkapiKey`) when no Okapi key is configured, and verify both grant constructors below return it before making any HTTP call

## 3. Client credentials grant

- [ ] 3.1 Implement `ClientCredentialsConfig` (`ClientID`, `ClientSecret`, `Endpoint`, `OkapiKey`, `Timeout time.Duration` defaulting to 30s when zero, optional `Scopes`) and `NewClientCredentialsSource(ctx context.Context, cfg ClientCredentialsConfig) (oauth2.TokenSource, error)` wrapping `clientcredentials.Config.TokenSource(ctx)`, passing `Endpoint.AuthStyle` (defaulted per 1.2) straight through to `clientcredentials.Config.AuthStyle`, and verify with a unit test against an `httptest.Server` stub token endpoint that a valid response yields a usable `*oauth2.Token`
- [ ] 3.2 Verify (unit test) that a token-endpoint error response (e.g. `invalid_client`/`bad_credentials`) surfaces as a returned `error` with no token, results in exactly one request to the stub server (no auto-retry with credentials moved to the form body), and that cancelling the constructor's `ctx` aborts an in-flight request
- [ ] 3.3 Verify (unit test) that every request to the stub token endpoint carries the `X-Okapi-Key` header from `2.1`/`2.2`, and that a stub server which never responds causes `Token()` to return an error once `Timeout` elapses rather than hanging indefinitely

## 4. Authorization code + PKCE grant

- [ ] 4.1 Implement `AuthorizationCodeConfig` (`ClientID`, `ClientSecret`, `Endpoint`, `RedirectURL`, `OkapiKey`, optional `Scopes`) and `NewAuthorizationCodeFlow(cfg AuthorizationCodeConfig) (*AuthorizationCodeFlow, error)`, generating and storing `state` and `code_verifier` on the returned value, and verify with a unit test that two calls produce different `state`/`code_verifier` pairs
- [ ] 4.2 Implement `(*AuthorizationCodeFlow) AuthCodeURL() string` embedding `client_id`, `redirect_uri`, `state`, and the PKCE `code_challenge`, and verify with a unit test that all four are present and the `code_challenge` matches `oauth2.S256ChallengeFromVerifier(flow.CodeVerifier)`
- [ ] 4.3 Implement `(*AuthorizationCodeFlow) Exchange(ctx context.Context, code, state string) (*oauth2.Token, error)` that rejects a `state` mismatch before calling the token endpoint, and verify with a unit test that a mismatched `state` returns an error and makes no HTTP call
- [ ] 4.4 Verify with a unit test against an `httptest.Server` stub token endpoint that a matching `state` plus a valid `code` yields a usable `*oauth2.Token`, and that the request includes the `code_verifier` and the `X-Okapi-Key` header
- [ ] 4.5 Verify (unit test) that exchanging an error response (e.g. `invalid_grant`) surfaces as a returned `error` with no token

## 5. Documentation

- [ ] 5.1 Add a `login/partner` row to `README.md`'s package table and a short "Partner API authentication" subsection under Authentication, describing both grants, the `X-Okapi-Key` requirement, and that it is independent of the self-vault login methods above it, and verify by re-reading the rendered section for accuracy against the implemented API
- [ ] 5.2 Add a package-level example (e.g. `login/partner/example_test.go`'s `Example`/`ExampleNewClientCredentialsSource`) demonstrating the client credentials flow end-to-end against a stub server, and verify `GOWORK=off go test ./login/partner/...` passes

## 6. Verification

- [ ] 6.1 Run `GOWORK=off go build ./...` and `GOWORK=off go test ./...` from the repo root and verify no existing package (`login`, `login/chrome`, `login/oauth`, `login/persistent`, `login/noop`, `v1`) is modified or broken
- [ ] 6.2 Run the project's lint target (`golangci-lint run ./...`, `GOWORK=off`) against the new package and verify it passes with the same rules as the rest of the repo
