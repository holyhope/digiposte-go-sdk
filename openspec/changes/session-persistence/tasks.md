## 1. Session store contract

- [ ] 1.1 Add a new package (e.g. `login/persistent`) with a `Session` struct (`Token *oauth2.Token`, `Cookies []*http.Cookie`), a `Store` interface (`Load(ctx) (*Session, error)`, `Save(ctx, *Session) error`), and an exported `ErrSessionNotFound` sentinel; verify with `go build ./...` and a doc comment on `Store` that states the security sensitivity of stored sessions (per design.md Decision 1 and Risks).
- [ ] 1.2 Add a package-local fake/in-memory `Store` implementation for tests only (not exported as a default), and unit tests covering `Load` returning `ErrSessionNotFound` when empty and round-tripping a saved `Session`; verify with `go test ./login/persistent/...`.

## 2. Fix `login/chrome` cookie seeding

- [ ] 2.1 In `chromeLogin.login`, before the first screen navigation, apply any cookies from `chrome.WithCookies` to the browser via the CDP network domain, scoped to the configured login URL's host; validate the cookie domain matches before seeding and return a clear error on mismatch (design.md Decision 4, Risk 2); verify with a unit/integration test asserting the cookies are present in the browser's cookie jar immediately after the first screen resolves.
- [ ] 2.2 Add a test (using the existing `login/chrome` test harness/mocked screens where possible, or the live-site test guarded the same way existing live tests are) that seeds a still-valid session's cookies and asserts the credentials/OTP/trusted-device screens are never invoked, only the final screen; verify by running that test and confirming no credential/OTP screen resolver logs a "Resolving screen" entry.
- [ ] 2.3 Add a test that seeds cookies that do not represent an authenticated session and asserts the normal credentials/OTP/trusted-device flow still runs unchanged; verify by running that test.

## 3. Session-resuming login decorator

- [ ] 3.1 Implement `persistent.NewMethod` factory type and `persistent.New(store Store, newMethod NewMethod) login.Method` per design.md Decision 2, implementing `login.Method.Login`; verify with `go vet ./...` and `go build ./...`.
- [ ] 3.2 Implement the resume path: `Login` calls `store.Load`, returns the loaded token/cookies directly (no factory call) when `token.Valid()`, treating both `ErrSessionNotFound` and any other `Load` error as "no session" rather than a fatal error (spec: "Fall back to interactive login..."); verify with a unit test asserting the factory is never invoked when a valid session is loaded, and a unit test asserting a `Load` error still results in a successful login via the factory.
- [ ] 3.3 Implement the fallback path: when no valid session was loaded, call `newMethod(cookies)` (with `cookies` nil when none were loaded) and delegate `Login` to the resulting method; verify with a unit test using a fake factory/method asserting the loaded cookies (or nil) are passed through.
- [ ] 3.4 Implement persistence of the outcome: after any successful `Login` (resumed or delegated), call `store.Save`; a `Save` error must be reported to the caller without turning an otherwise-successful login into a failure (spec: "Save failure does not fail the login") - decide and document the exact error-reporting shape (e.g. `errors.Join` alongside a nil login error is not applicable since login succeeded; likely a distinct returned error value or logged warning - resolve during implementation, consistent with design.md's decision to keep `Login`'s success independent of `Save`'s outcome); verify with a unit test asserting a `Save` error does not cause `Login` to return an error while a `Save` success is later observable (e.g. via the fake store's recorded calls).
- [ ] 3.5 Verify the "no store configured" and "store optional" behavior described in the spec is naturally satisfied by requiring a non-nil `Store` at construction (`persistent.New` returns an error or panics on a nil store) or, if a nil store is allowed as a pass-through, add an explicit test proving no `Load`/`Save` call happens and behavior is a plain delegate to `newMethod(nil)`; verify with a unit test for whichever behavior is chosen.

## 4. Integration and examples

- [ ] 4.1 Wire `login/chrome` as a `persistent.NewMethod` factory in a runnable example mirroring `login/oauth/tokensource_example_test.go` (i.e. `persistent.New(store, func(cookies []*http.Cookie) (login.Method, error) { return chrome.New(append(baseOpts, chrome.WithCookies(cookies))...) })` composed with `login/oauth.TokenSource` and `oauth2.ReuseTokenSource`); verify by running `go test ./login/... -run Example`.
- [ ] 4.2 Update the README (or package doc comments if the README doesn't cover login construction in detail) to document the new package, the `Store` contract, and the explicit "no default store, bring your own storage" decision, including the security note from design.md; verify by a manual read-through confirming no plaintext-storage recommendation is implied.

## 5. End-to-end verification

- [ ] 5.1 Run the full test suite (`go build ./...`, `go vet ./...`, `golangci-lint run`) and confirm zero issues; verify by inspecting command output.
- [ ] 5.2 Manually validate the resume path against a real (non-rate-limited) Digiposte session: obtain a session once via `chrome.New`, save it through a throwaway `Store` implementation, then confirm a second run resumes without launching Chrome's interactive screens; verify by observing no credentials/OTP/trusted-device screen logs on the second run.
