## Context

See proposal.md - Why. Two existing facts shape this design:

- `login.Method` is a narrow interface (`Login(ctx, creds) (*oauth2.Token, []*http.Cookie, error)`), and the SDK already composes login methods by decoration: `login/oauth.TokenSource` wraps a `login.Method` to make it an `oauth2.TokenSource`, and callers typically wrap that again in `oauth2.ReuseTokenSource` for in-process caching (see `login/oauth/tokensource_example_test.go`). This change adds one more decorator at the same layer, so it composes with what already exists instead of replacing it.
- `login/chrome` already has a `WithCookies` option whose value (`chromeLogin.cookies`) is never read. There is no other hook today for getting cookies into the browser before it navigates.

## Goals / Non-Goals

**Goals:**
- Let a caller skip the interactive Chrome flow entirely when a still-valid session is available.
- Let a caller who has a still-authenticated (but not `oauth2.Token`-fresh) site session reuse it by seeding its cookies, instead of re-entering credentials/OTP.
- Keep the addition composable with the existing `login/oauth.TokenSource` + `oauth2.ReuseTokenSource` stacking pattern.

**Non-Goals:**
- No reactive re-authentication when a resumed session turns out to be rejected by a later `v1` API call (e.g. a 401). Callers still get that failure today; handling it is a separate, later change.
- No default `Store` implementation (per the confirmed decision: interface only, caller supplies storage).
- No change to how `v1.TokenSource` renews a bearer token from cookies, or to any `v1/*` resource client.
- No attempt to fix the underlying DOM-selector fragility of the interactive flow itself - this change reduces how often that flow runs, not its inherent fragility when it does run.

## Decisions

### 1. A `Store` interface keyed on token + cookies, with a sentinel "not found" error
```go
type Session struct {
    Token   *oauth2.Token
    Cookies []*http.Cookie
}

type Store interface {
    Load(ctx context.Context) (*Session, error)
    Save(ctx context.Context, session *Session) error
}

var ErrSessionNotFound = errors.New("session not found")
```
`Load` returns `ErrSessionNotFound` (checked with `errors.Is`) when nothing is stored yet, rather than `(nil, nil)`. This follows the repo's existing convention of exported sentinel errors (e.g. `oauth.ErrNoTokenSources`) instead of an ambiguous double-nil.

Alternative considered: a `(session *Session, found bool, err error)` return shape. Rejected - three-way returns are more error-prone to consume correctly than a sentinel error, and `errors.Is` composes with `errors.Join`, which the codebase already uses.

### 2. The decorator takes an inner-method *factory*, not a ready-built `login.Method`
```go
type NewMethod func(seedCookies []*http.Cookie) (login.Method, error)

func New(store Store, newMethod NewMethod, opts ...Option) login.Method
```
`opts` configures cross-cutting behavior such as save-error reporting - see Decision 3.

On `Login`, the decorator: loads from `Store`; if the token is valid, returns it directly (no factory call at all); otherwise calls `newMethod(session.Cookies)` (with `session.Cookies` nil if nothing was stored or load failed) to obtain a `login.Method` and delegates to it; then saves whatever it returns.

Alternative considered: accept an already-constructed `login.Method` and extend `login.Method` with an optional `interface{ WithSeedCookies([]*http.Cookie) login.Method }` that `login/chrome`'s method would implement, type-asserted by the decorator. Rejected - it requires a new cross-package optional-interface convention for something a plain closure already does; the factory shape lets a caller write `func(cookies []*http.Cookie) (login.Method, error) { return chrome.New(append(baseOpts, chrome.WithCookies(cookies))...) }` with no new interface at all, and keeps `login.Method` itself unchanged.

### 3. Always attempt `Save` after a successful `Login`, on both the resumed and interactive paths
Simpler single code path ("obtain a session, then save it") beats special-casing "don't save if nothing changed." A `Save` on an unchanged, already-valid session is a harmless no-op for any reasonable `Store` implementation. A `Save` error must not fail an otherwise-successful `Login`. Concretely: `New` accepts a variadic `opts ...Option`, and `WithSaveErrorHandler(func(error))` registers a callback invoked with a `*SaveError` wrapping the underlying `Save` error; if no handler is configured, the default logs the error via the standard `log` package. Either way, `Login` still returns the obtained token and cookies with a nil error.

### 4. Seed cookies into the browser via CDP `Network.setCookie`, scoped to the login domain, before the first navigation
This is the fix for `WithCookies` being a no-op today: in `chromeLogin.login`, before `resolve(ctx, &firstScreen{...})` runs, apply any configured cookies to the browser's network domain for `c.url`'s host. No changes are needed to the existing screen-resolution state machine - `Screens.Resolve` already reacts to whatever screen the live site renders. If the seeded cookies represent a still-authenticated session, the site itself will skip straight past the credentials/OTP/trusted-device screens (they simply won't be the screen shown), and the resolver's existing final-screen match takes over unchanged. If the cookies are stale, the site shows the normal screens and the flow proceeds exactly as it does today with no seeded cookies at all.

### 5. Validity check for the fast (no-Chrome) path is `oauth2.Token.Valid()` only
No live network probe before deciding to skip Chrome - `Token.Valid()` (expiry-based, with the library's built-in skew buffer) is the same check `oauth2.ReuseTokenSource` already relies on elsewhere in this codebase. See Risks below for the resulting limitation.

## Risks / Trade-offs

- **[Risk]** A resumed token can be `Valid()` by expiry yet already rejected server-side (revoked session, changed password, or the kind of account-level block already seen in production) → **Mitigation**: out of scope here (see Non-Goals); document the limitation on `Store`/the decorator's doc comment. A caller that observes repeated downstream 401s can clear its `Store` to force a fresh interactive login.
- **[Risk]** Seeded cookies with an unexpected domain/path could be rejected or silently ignored by `Network.setCookie` → **Mitigation**: validate the cookie domain against the configured login URL before attempting to seed, and surface a clear error rather than proceeding silently into a login attempt that cannot succeed.
- **[Risk]** A `Store` is equivalent to storing live account credentials; a careless implementation (plaintext file with wide permissions, accidental logging, checked into version control) is a real account-security exposure → **Mitigation**: ship no default `Store` (confirmed decision), and make this sensitivity explicit and prominent in the interface's doc comment.
- **[Trade-off]** The factory-function shape (`NewMethod`) is a slightly different integration pattern than the existing `oauth.TokenSource{LoginMethod: ...}` decorator, which takes an already-built method → **Mitigation**: keep the factory signature minimal (one parameter) and ship a runnable example mirroring the existing `tokensource_example_test.go`, showing the one-line change needed to adopt it.
- **[Risk]** On the fallback path, the decorator calls `newMethod` (the interactive/Chrome factory) directly once the stored token is invalid or expired; it does not first attempt a cheaper cookie-based HTTP token renewal, even though `v1.NewAuthenticatedClient` already composes exactly that ("try `v1.TokenSource` over plain HTTP via cookies before falling back to an interactive `login.Method`", see `v1/client.go`'s `oauth.CombinedTokenSources{tokenSource, ...}`). A caller using `login/persistent` on its own therefore launches Chrome on every token expiry even when the stored cookies would have renewed over HTTP. → **Mitigation**: known limitation for this change; `login/persistent` operates at the `login.Method` layer and has no dependency on `v1` (avoiding an import cycle), so it cannot itself call `v1.TokenSource`. Callers who want cookie-based renewal before Chrome should compose `login/persistent` inside a `v1.Config{PreviousSession: ..., LoginMethod: ...}` rather than relying on `login/persistent` alone. A follow-up change should evaluate lifting `NewMethod`'s fallback role, or `login/persistent` itself, into (or alongside) `v1.NewAuthenticatedClient`'s existing `CombinedTokenSources` composition, so renewal-before-Chrome and persistence-across-restarts aren't two disconnected mechanisms.

## Migration Plan

Purely additive - no existing exported API changes behavior. Callers who never use the new package or `chrome.WithCookies` see no difference. No rollback concerns beyond a normal revert, since nothing existing is modified except making a previously-inert option (`WithCookies`) functional.

## Open Questions

- Exact strictness of cookie domain/path validation when seeding (reject on any mismatch vs. filter mismatched cookies and proceed with the rest) - an implementation detail resolvable during coding/review without changing the spec or task breakdown.
- Final package name for the new decorator (e.g. `login/persistent` vs `login/session`) - naming only, does not affect behavior or the task breakdown.
