## Why

`login/chrome`'s screen logic (`CurrentPageMatches`/`Do` for the privacy, credentials, OTP, trusted-device, and final screens) is only exercised two ways today: `screens_internal_test.go`'s synthetic `Screen` fakes, which never touch chromedp or any real DOM at all, and `chrome_test.go`'s Ginkgo specs, which drive a real Chrome browser against the live `digiposte.fr` / La Poste SSO site with real credentials (`DIGIPOSTE_USERNAME`/`PASSWORD`/`OTP_SECRET`) and self-skip otherwise. That leaves no way to exercise the actual chromedp selectors (`#username`, `#submit-button`, `#otpCode`, `#save-trusted-device-form`, ...) against realistic markup without a live account, a live browser, and minutes of wall-clock time per run. It also carries real operational risk: repeated automated logins against the live SSO can get rate-limited or blocked outright (a `403 Forbidden` was hit today after a day of debugging attempts), which stops all screen-level testing dead regardless of whether the code under test is correct.

We should capture ("dump") the real HTML of each login screen to local fixture files once, and add a local mock HTTP server driven by those fixtures so `login/chrome`'s screen-selector logic can be exercised repeatedly, quickly, deterministically, and offline - without a live account, a live browser session against the real site, or rate-limit risk.

## What Changes

- Add a test-only capture helper (an exported symbol in a new `login/chrome/export_test.go`, never part of the public API) that records every CDP network response - documents and assets (CSS/JS/images) alike - while a real login session runs once, and saves each one under `login/chrome/testdata/<hostname>/...`, mirroring its real path under the hostname that served it.
- Add a lightweight local HTTP server (`net/http/httptest`-based) that resolves each real hostname involved in the flow to itself via Chrome's `--host-resolver-rules`, and dispatches incoming requests by `Host` header + path to that hostname's captured directory tree, serving pages and assets byte-for-byte; any hostname without a captured tree is refused rather than passed through. The final page's captured fixture is amended with a small hand-authored script seeding the `sessionStorage` keys (`access_token`, `app_expires_at`) and cookies that `finalScreen.Do` reads, since a real page never contains those verbatim.
- Add Ginkgo specs (in `login/chrome`) that drive `chrome.New(chrome.WithURL(mockServer.URL), ...)` against this mock server instead of the live site, verifying each screen's selectors and interactions (fill/click/submit, "skip OTP", "don't trust this device", token/cookie extraction) still work - without `DIGIPOSTE_USERNAME`/`PASSWORD`/`OTP_SECRET` or a live browser session against the real site.
- These new specs run unconditionally in CI (no live secrets, no live account needed), unlike `chrome_test.go`'s existing specs, which stay skipped without real credentials.
- No changes to `login/chrome`'s public API, options, or production runtime behavior: the same `Screen` implementations, selectors, and `Do()`/`CurrentPageMatches()` logic run unmodified, just against local fixture content instead of the real site.

## Capabilities

This is testing/tooling infrastructure for `login/chrome`: it adds a way to test the package faster and without hitting the live site, but does not change what `login/chrome` does at runtime or any spec-level behavior. No capability specs change; see `skip_specs: true` in `.openspec.yaml`.

### New Capabilities

(none)

### Modified Capabilities

(none)

## Impact

- New `login/chrome/testdata/<hostname>/...` fixture files (pages and assets), captured once from a real login session, organized per real hostname.
- New capture tooling (test-only, `export_test.go`) to (re)generate those fixtures when the real site's markup changes (e.g., selector drift, as already happened twice for the OTP and trusted-device screens).
- New mock HTTP server test helper in `login/chrome`'s test code, using a Chrome host-resolver override and Host-header dispatch to serve each hostname's captured tree.
- No new public API: the capture helper lives only in a `_test.go` file and is invisible to normal consumers of the module.
- New Ginkgo specs in `login/chrome` exercising every screen against the mock server; these run in CI unconditionally, unlike the existing live-account specs.
- `login/chrome`'s existing live specs in `chrome_test.go` are unchanged (still gated on env vars) and remain the true end-to-end check; the new fixture-based specs are a fast, deterministic complement, not a replacement.
- No changes to `login/chrome`'s production code paths, public API, or options.
