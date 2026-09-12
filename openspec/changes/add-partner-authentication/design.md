## Context

See `proposal.md` - Why/What Changes for motivation and scope. Relevant existing shape:

- `login.Method`/`login.Credentials` (in `login/`) model a single interactive end-user login that returns an `*oauth2.Token` *and* `[]*http.Cookie` - cookies matter there because the self-vault site (`secure.digiposte.fr`) is session-cookie-based underneath its bearer token. The Partner API is a pure OAuth2/REST API with no cookie concept, so reusing `login.Method`'s shape would force an unused `[]*http.Cookie` return everywhere - this is why the proposal keeps `login/partner` as its own subpackage instead of implementing `login.Method`.
- The exact sandbox/production base URLs for the Partner API's `/authorize` and `/token` endpoints are not consistently documented in public sources (some list `.../v3/token`, others `.../v3/oauth/token`, and both sandbox and production have distinct hostnames per environment). Each registered partner gets their own OKAPI space and Swagger contract, so this SDK should not hardcode guessed URLs.
- `golang.org/x/oauth2` (already a dependency) provides `clientcredentials.Config` for the `client_credentials` grant and, on `oauth2.Config`, `AuthCodeURL`/`Exchange` plus PKCE helpers (`GenerateVerifier`, `S256ChallengeOption`, `VerifierOption`) for the `authorization_code` grant - no new third-party dependency is needed for either grant.

## Goals / Non-Goals

**Goals:**
- Provide two small, independent constructors in `login/partner` - one per grant type - each producing a working `oauth2.TokenSource`/token, with the `X-Okapi-Key` header applied to every token request either makes.
- Let the caller supply the exact endpoints (sandbox or production, or a partner-specific OKAPI URL) rather than the SDK guessing at unverified hardcoded values.
- Keep the authorization-code flow's CSRF `state` check and PKCE `code_verifier` handling inside the package, so callers don't have to reimplement OAuth2 plumbing correctly themselves.

**Non-Goals:**
- Implementing any Partner API resource endpoint (memberships/adhésions, document deposit/retrieval) - only authentication. A follow-up change can build a partner API client on top of the token this package produces.
- Persisting or refreshing tokens across process restarts (unlike `login/persistent`, which exists for the self-vault flow) - callers needing that can wrap the returned `oauth2.TokenSource` with `oauth2.ReuseTokenSource` themselves, same as any other `oauth2.TokenSource`.
- Running an HTTP server/handler to receive the `authorization_code` redirect - the caller's own web server must receive the callback and pass the `code`/`state` query parameters into this package's `Exchange` call.

## Decisions

- **Separate `login/partner` package, not a `login.Method` implementation.** Rationale in Context above. Alternative considered (making `login/partner` implement `login.Method` by returning `nil` cookies) was rejected: it would let a partner token be silently fed into `login/persistent` or `v1.Config`, which only make sense for self-vault sessions, and would suggest a compatibility that doesn't exist.
- **Two independent constructors, no shared "Method" abstraction between the two grants.** `partner.NewClientCredentialsSource(cfg ClientCredentialsConfig) oauth2.TokenSource` and `partner.NewAuthorizationCodeFlow(cfg AuthorizationCodeConfig) *AuthorizationCodeFlow` (with `AuthCodeURL() string` and `Exchange(ctx, code, state string) (*oauth2.Token, error)` methods). They have different shapes (one is stateless and immediately usable, the other is stateful across a redirect round trip) and forcing a common interface would only hide that difference behind an awkward abstraction with no real caller benefit.
- **Caller supplies endpoints; the SDK does not hardcode sandbox/production URLs.** Both `ClientCredentialsConfig` and `AuthorizationCodeConfig` take an `Endpoint` (`AuthURL`, `TokenURL` - mirroring `oauth2.Endpoint`) field the caller fills in from their own OKAPI/Swagger contract, rather than the package guessing at URLs pulled from inconsistent public documentation. Alternative considered: ship `partner.SandboxEndpoint`/`partner.ProductionEndpoint` constants - rejected for now because the exact values could not be confirmed with certainty; can be added later as convenience constants once confirmed, without breaking this design (an additive follow-up, not a redesign).
- **`X-Okapi-Key` applied via a shared `http.RoundTripper` wrapper**, not by requiring callers to set headers manually. `login/partner` builds an `*http.Client` whose `Transport` injects `X-Okapi-Key` on every request and installs it into the `oauth2` library's request context (`context.WithValue(ctx, oauth2.HTTPClient, client)`) before calling into `clientcredentials.Config.Token`/`oauth2.Config.Exchange`. This guarantees the header is present on every token request from both grants without duplicating header-setting logic in two places, and the same transport can be reused by a later change's resource-request client.
- **The authorization-code flow owns `state` and `code_verifier` generation and validates `state` itself.** `NewAuthorizationCodeFlow` generates both once (via `oauth2.GenerateVerifier` and a package-level `generateState` helper of equivalent strength) and stores them on the returned `*AuthorizationCodeFlow` value; `AuthCodeURL()` embeds them, and `Exchange` rejects a mismatched `state` before attempting the token exchange. Alternative considered: let the caller pass in their own `state`/`verifier` - rejected as the default because it is easy to get wrong (e.g. reusing a `state` across requests) or forget entirely; the underlying `oauth2.Config.AuthCodeURL`/`Exchange` calls are still reachable if a caller has an unusual need to supply their own values in a follow-up change.

## Risks / Trade-offs

- [Unverified Partner API endpoint URLs] → Mitigated by never hardcoding them: `Endpoint` is a required, caller-supplied field, so a wrong guess in this SDK can't cause runtime failures - the caller is responsible for getting it from their own OKAPI Swagger contract.
- [`*AuthorizationCodeFlow` holds `state`/`code_verifier` in memory, so it must survive between `AuthCodeURL()` and the later `Exchange()` call - a problem for multi-instance/stateless deployments where the process that builds the URL may not be the one that receives the callback] → Mitigate by exposing the generated `state`/`code_verifier` as readable fields on the returned flow value, so a caller that cannot keep the Go value alive across the round trip (e.g. a horizontally-scaled web server) can persist those two strings themselves (e.g. keyed by `state` in their own session store) and reconstruct an equivalent flow value before calling `Exchange`. Document this explicitly rather than assuming single-process affinity.
- [Two grants with different shapes (stateless token source vs. stateful multi-step flow) could read as an inconsistent API] → Mitigated by keeping both constructors' names and doc comments explicit about what each is for, and by not forcing a shared interface that would hide the difference (see Decisions above).

## Migration Plan

Purely additive - no migration needed. No existing package, type, or behavior changes; nothing to roll back beyond reverting the new `login/partner` package and its README section if needed.
