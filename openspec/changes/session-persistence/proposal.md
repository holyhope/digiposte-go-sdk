## Why

Establishing a Digiposte session today always requires driving a real (undetected) Chrome browser through the site's login screens (privacy consent, credentials, OTP, trusted device). This is the SDK's most fragile and highest-risk path: DOM selectors drift as the live site changes, and rapid/automated re-authentication attempts have already triggered an account-level rate limit (HTTP 403) in production. Yet the SDK already mints a fresh bearer token from existing session cookies via a plain HTTP call (`v1.TokenSource`, hitting `/rest/security/token`) without ever touching Chrome again - Chrome is only strictly required to *establish* a session, not to keep using one. Nothing in the SDK currently persists a session across process restarts, so every run pays the full Chrome cost and risk even when a still-valid session already exists. `login/chrome` even has an unfinished `WithCookies` option (the field is set but never read), showing this gap was already anticipated but not completed.

## What Changes

- Add a `Store` interface for saving and loading a previously obtained session (`*oauth2.Token` + `[]*http.Cookie`), matching the SDK's existing pluggable-interface style (`login.Method`, `login.Option`).
- Add a `login.Method` decorator that wraps any inner method (e.g. `chrome.New(...)`): before delegating, it loads a session from the `Store`; if the loaded token is still valid, it returns it directly with no interactive login at all; otherwise it seeds the loaded cookies into the inner method (see below) and falls back to delegating. After any successful login (fresh or resumed), it saves the resulting token + cookies back to the `Store`.
- Fix `login/chrome`'s `WithCookies` option so seeded cookies actually reach the browser before navigation (the field is currently set but never applied). When the seeded cookies represent a still-valid site session, the existing screen-resolution flow should recognize the user as already authenticated and skip the credentials/OTP/trusted-device screens entirely, reaching the final screen without any DOM scripting of those screens.
- No default `Store` implementation ships in this change; the SDK only defines the contract and the wiring that uses it. Callers provide their own storage (file, keychain, secrets manager, etc.) and are responsible for protecting it, since a stored session is equivalent to live account credentials.
- **BREAKING**: none anticipated - this is purely additive (new interface, new decorator, a bug fix to a currently-nonfunctional option). Existing callers who never set `chrome.WithCookies` see no behavior change.

## Capabilities

### New Capabilities
- `login/session-persistence`: defines the `Store` contract, the session-resuming `login.Method` decorator, and the requirement that a seeded, still-valid session lets the chrome login method skip its interactive screens instead of re-running the full credentials/OTP/trusted-device flow.

### Modified Capabilities
(none - `login/chrome`'s existing screen-resolution behavior has no prior spec; this change only adds new, previously-undocumented behavior to it)

## Impact

- New code: a session-store interface and a decorator implementing `login.Method`, likely in a new `login/` subpackage (exact location decided in design.md).
- Modified code: `login/chrome` - wire the existing `cookies` field into the browser session (via CDP) before the first screen navigates, so a valid seeded session is recognized as already authenticated.
- No changes to `v1/*` (document/folder/share/token clients) or to on-the-wire API calls - this only affects how a session is obtained/reused before those clients are used.
- No new third-party dependencies (no bundled storage backend).
- Directly reduces exposure to the rate-limiting/DOM-drift risk described above by making the interactive Chrome flow the exception rather than the rule for callers who opt in to a `Store`.
