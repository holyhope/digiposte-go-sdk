## Why

`login/chrome`'s screen resolver polls the page at `WithRefreshFrequency` and, on each match, runs the matched screen's `Do()` (its Chrome automation) inside a context whose deadline is that *same* duration (`screens.go`'s `run` uses `context.WithTimeout(ctx, s.refreshFrequency)` around `resolve`). Fast, single-step screens (a lone click, like `privacyScreen`) fit comfortably inside any reasonable frequency, but multi-step screens - `credentialsScreen` (wait, click, clear, type, wait, click, clear, type, click submit), `otpScreen`, and `trustedDeviceScreen` - routinely need more wall-clock time against a real page than a tight polling interval allows. When a caller (or a test, to keep runs fast) sets a short `WithRefreshFrequency`, the multi-step screen's `Do()` is killed mid-sequence by `context.DeadlineExceeded` before it can submit, the next poll tick re-matches the same screen and restarts the sequence from scratch, and the login effectively never completes - it loops forever until the outer `WithTimeout` (or the caller's context) finally expires. This was hit directly running the v1 integration suite with `WithRefreshFrequency(500 * time.Millisecond)`: the credentials screen never finished, and a workaround of raising that one duration to 10s made the suite pass again - but it does so by also slowing down every other screen's polling cadence, which is not what the caller intended and won't scale if any screen someday needs even longer.

## What Changes

- Add a new, distinct option (e.g. `WithScreenTimeout`) to `login/chrome` that sets how long a single `resolve` attempt (a matched screen's `Do()`) is allowed to run, independently of `WithRefreshFrequency` (which continues to control only how often `CurrentPageMatches` is polled).
- In `screens.go`'s `run`, use the new screen-execution timeout instead of `s.refreshFrequency` for the `context.WithTimeout` wrapped around `resolve(ctx, screen)`.
- Give the new option a default that comfortably covers the existing multi-step screens against a real page (independent of whatever `WithRefreshFrequency` a caller picks), so today's callers who never set the new option see no regression, and callers who set a short `WithRefreshFrequency` (for fast polling) no longer starve `Do()` of the time it needs to finish.
- Verify `v1/v1_suite_test.go`'s existing short `WithRefreshFrequency(500 * time.Millisecond)` now reliably completes the full login flow once execution time is no longer tied to it (previously, only a manual, uncommitted local workaround of raising it to `10 * time.Second` made the suite pass, and even that shared-knob approach masked rather than fixed the underlying conflation).
- No change to the set of screens, their selectors, or their individual `Do()` logic.

## Capabilities

### New Capabilities
- `login/screen-resolution-timing`: defines that a screen's polling frequency (how often the page is checked) and its execution timeout (how long a matched screen is given to complete its automation) are independent, separately configurable durations, and that the execution timeout must be long enough for the login flow's built-in multi-step screens to complete against a real page under the default configuration.

### Modified Capabilities
(none - there is no existing spec covering `login/chrome`'s screen-resolution polling/timeout behavior; this only adds new, previously-undocumented behavior)

## Impact

- Modified code: `login/chrome/screens.go` (execution timeout source), `login/chrome/options.go` (new option), `login/chrome/login_method.go` (new default constant, wiring into `chromeLogin`).
- Modified test (verification only, no behavior-changing edit expected): `v1/v1_suite_test.go` - its existing `WithRefreshFrequency(500 * time.Millisecond)` should now reliably pass once execution time is decoupled from it.
- No change to `v1/*` document/folder/share/token clients, no new third-party dependencies, no breaking change - existing callers who only use `WithRefreshFrequency` keep their current polling behavior, and gain a working (non-looping) login by default.
