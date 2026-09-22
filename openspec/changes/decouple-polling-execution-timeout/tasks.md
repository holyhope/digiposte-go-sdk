## 1. Screen execution timeout option

- [ ] 1.1 Add `WithScreenTimeout(timeout time.Duration) login.Option` to `login/chrome/options.go`, mirroring the existing `WithTimeout`/`WithRefreshFrequency` pattern (a `Validatable` struct with `Apply` setting a new `chromeLogin.screenTimeout` field, and `Validate` rejecting non-positive durations with a `login.InvalidOptionError`), and verify with a new unit test covering both the valid-apply and non-positive-rejection paths (mirroring the existing `WithTimeout`/`WithRefreshFrequency` tests).
- [ ] 1.2 Add `screenTimeout time.Duration` to the `chromeLogin` struct (`login/chrome/chrome.go`) and add a `DefaultScreenTimeout = 30 * time.Second` constant (see design.md - Decisions for the headroom rationale) plus its default wiring in `newChromeLogin` (`login/chrome/login_method.go`), and verify by inspection that a `chromeLogin` built with no `WithScreenTimeout` option has `screenTimeout == DefaultScreenTimeout`.

## 2. Decouple polling from execution in the resolver

- [ ] 2.1 Add a `screenTimeout time.Duration` field to `Screens` (`login/chrome/screens.go`), thread `c.screenTimeout` into it from `resolveLogin` (`login/chrome/chrome.go`) alongside the existing `refreshFrequency` field, and change `run`'s `context.WithTimeout(ctx, s.refreshFrequency)` to use `s.screenTimeout` instead - verify by reading the diff that `refreshFrequency` is now used only for the polling `time.NewTicker`, not for any `resolve` call's deadline.
- [ ] 2.2 Add or update a unit test in `login/chrome` (e.g. in `screens.go`'s test file) that exercises a screen whose `Do()` takes longer than a short configured `refreshFrequency` but less than `screenTimeout`, and verify the screen still resolves successfully instead of hitting `context.DeadlineExceeded`.

## 3. Restore fast polling in the v1 suite

- [ ] 3.1 Confirm `v1/v1_suite_test.go` still has its existing `chrome.WithRefreshFrequency(500 * time.Millisecond)` (no committed edit needed - only a local, uncommitted `10 * time.Second` workaround was ever tried, and it should not be merged) and leave `WithScreenTimeout` unset (relying on the new default), then verify by running `go test ./v1 -run TestV1 -v -args --ginkgo.focus="Should create a document"` (with `.env` sourced) that the full login flow (privacy, credentials, OTP, trusted device, final) completes and the spec passes.

## 4. Full verification

- [ ] 4.1 Run `go test ./...` and verify all packages pass, confirming the change is additive and does not break existing `login/chrome` behavior for callers who set neither new option.
