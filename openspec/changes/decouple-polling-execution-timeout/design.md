## Context

`Screens.run` (`login/chrome/screens.go`) polls each screen's `CurrentPageMatches` on a ticker fed by `refreshFrequency`, and on a match wraps the call to `resolve(ctx, screen)` in `context.WithTimeout(ctx, s.refreshFrequency)` - the exact same duration. `refreshFrequency` is set from a single option, `WithRefreshFrequency` (`login/chrome/options.go`), with `DefaultRefreshFrequency = 1500 * time.Millisecond` (`login/chrome/login_method.go`). There is currently no separate knob for how long a matched screen's `Do()` may run. See `proposal.md - Why` for how this surfaced (the v1 integration suite deadlocking with a 500ms `WithRefreshFrequency`).

## Goals / Non-Goals

**Goals:**
- Give screen execution its own timeout, independent of the polling ticker interval.
- Keep the default behavior for existing callers who only ever set `WithRefreshFrequency` (or nothing) working at least as well as today's `DefaultRefreshFrequency` (1500ms) execution budget - ideally better, since some multi-step screens already need more than that against a real page.
- Keep the change additive and localized to `login/chrome`.

**Non-Goals:**
- Not changing per-screen selectors, retry/backoff strategy, or adding per-screen-type timeouts. One execution timeout applies to whichever screen is currently being resolved.
- Not changing the outer `WithTimeout` (overall login deadline) semantics.
- Not adding a moving/adaptive timeout (e.g. longer on retry) - out of scope for this change.

## Decisions

**New option name and shape: `WithScreenTimeout(timeout time.Duration) login.Option`.**
Mirrors the existing `WithTimeout`/`WithRefreshFrequency` option pattern exactly (a `Validatable` struct with an `Apply` and a `Validate` that rejects non-positive durations), so it's a drop-in stylistic match with `login/chrome/options.go`'s existing options rather than a new pattern.
Alternative considered: overload `WithTimeout` (the overall login deadline) to also seed a fraction of itself as the per-screen timeout. Rejected - `WithTimeout` bounds the entire flow (all screens plus navigation), and deriving a per-screen budget from it (e.g. dividing by screen count) would be an implicit, surprising coupling; an explicit, separate option keeps both concerns independently reasoned about.

**New field: `chromeLogin.screenTimeout`, threaded into `Screens.screenTimeout` (`screens.go`), replacing the use of `s.refreshFrequency` inside `run`'s `context.WithTimeout` call.**
`refreshFrequency` keeps its existing, sole meaning: the ticker interval for `CurrentPageMatches` polling. Nothing else about `Screens.run`'s control flow changes - it is a one-line source swap for the timeout duration.

**Default value: `DefaultScreenTimeout = 30 * time.Second`.**
The manual run that motivated this change observed the credentials screen's ~10 sequential CDP round-trips (wait/click/clear/type ×2 plus a final click) completing quickly, but the OTP screen took ~10s including one internal retry, and the trusted-device screen ~4s. A default set close to that observed 10s ceiling would leave essentially no headroom against ordinary scheduling or network jitter, and a single slow tick would recreate the exact timeout/retry loop this change exists to prevent - failing the very requirement (built-in screens complete in a single attempt under normal conditions) this capability adds. `30 * time.Second` gives roughly 3x headroom over the worst observed case (OTP with a retry) while still being well inside the outer `WithTimeout` login deadline (3 minutes in the v1 suite). This does not affect existing callers' polling cadence - only how long a matched screen now gets to actually finish, which was previously silently capped by whatever `WithRefreshFrequency` they had chosen.
Alternative considered: default it to whatever `WithRefreshFrequency` is set to, preserving today's exact behavior unless a caller opts in to a longer value. Rejected as the "default" that reproduces the bug this change exists to fix - a caller who deliberately picks a short polling frequency (e.g. for a fast test double or a mocked page) would still get starved automation on any real, slower page unless they also remember to set the new option; a fixed, generous default removes that footgun while still being fully overridable.

**`v1/v1_suite_test.go`: keep its existing `WithRefreshFrequency(500 * time.Millisecond)` unchanged, and do not set `WithScreenTimeout` explicitly.**
The committed suite already uses `500ms` - the `10 * time.Second` value was only a manual, uncommitted local workaround explored while diagnosing this bug, never merged. Once execution time is no longer tied to the polling interval, that `500ms` setting (as originally intended by its "Reduce the test duration" comment) should reliably pass on its own: it continues to control only how snappy detection of a screen match is, while the new default screen-execution timeout covers the multi-step screens' actual automation time.

## Risks / Trade-offs

- [A single global execution timeout is still shared by every screen type, so a future screen slower than 10s would need either a bigger default or the not-yet-built per-screen override] → Mitigation: the option is a simple, explicit knob a caller can already raise; per-screen timeouts can be added later as a non-breaking extension (e.g. an optional `Screen`-level override) if a real need arises - not speculatively building it now.
- [Raising the effective execution budget for slow/misbehaving screens means a single stuck screen can now block the outer login for longer per attempt (up to `DefaultScreenTimeout` instead of the previous, shorter `refreshFrequency`-derived budget) before the poll loop retries it] → Mitigation: the outer `WithTimeout` (overall login deadline) is unchanged and still bounds total wall-clock time; a stuck screen still eventually surfaces as a timeout error to the caller, just with a longer, more realistic per-attempt budget instead of one that was already too short to succeed.

## Migration Plan

Purely additive: new option, new default, no signature changes to existing exported functions. Existing callers who never set `WithRefreshFrequency` or `WithScreenTimeout` get the new, longer default execution budget automatically - the only observable change for them is that logins which previously silently looped/timed out on slow screens can now succeed. No rollback concerns beyond a plain revert if the new default ever proves too generous for some caller.
