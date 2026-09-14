## Context

`.github/workflows/test.yml` has one job (`tests`) with one test step, `go test -v ./v1/... -ginkgo.v`, and `push`/`pull_request` triggers gated on `paths: ['**.go', '!**_test.go', 'v1/**']`. See `proposal.md` - Why for the coverage gap this leaves for `login/partner`. Separately, the real Digiposte Partner API sandbox shares its base URL with production (`https://api.laposte.fr/digiposte/v3/token`, confirmed manually) and accepts a publicly documented test `client_id`/`client_secret` (`okapi`/`6mstVSvc38wq`); only the Okapi key passed via `X-Okapi-Key` is a real, non-public credential.

## Goals / Non-Goals

**Goals:**

- Make a `login/partner` test failure fail CI, the same way a `v1` test failure does today.
- Keep the push/PR-triggered step free of any `DIGIPOSTE_*` secrets/vars - `login/partner`'s ordinary tests only use `httptest.Server` stubs, and the new live spec must self-skip (not fail) when no live secret is present.
- Trigger on a `login/partner`-only test-file change, which the current path filter drops.
- Run one live spec against the real Partner API sandbox on a low, predictable cadence (weekly, alongside the existing `schedule` trigger), so a Digiposte-side API change surfaces automatically instead of only when someone happens to test manually.

**Non-Goals:**

- Closing the same CI gap for `login/chrome`, `login/oauth`, or `login/persistent` (explicitly out of scope per proposal.md - Impact).
- Changing anything about how `v1`'s tests run, or the weekly `schedule` trigger's cron cadence.
- Adding a separate GitHub Actions job. Two extra steps in the existing `tests` job are enough here; a second job would duplicate the checkout/setup-go steps.
- Running the live spec on every push/PR, like `v1`'s live tests do - that trades quota/flake risk for signal this change doesn't need; a weekly cadence is enough to catch a sandbox change.

## Decisions

- **New step in the existing job, not a new job.** The existing job already has Go set up; adding `- name: Test login/partner` \ `run: go test -v ./login/partner/... -ginkgo.v` after the current `Test` step reuses that setup and keeps a single place to look for test results. Alternative considered: a separate `partner-tests` job running in parallel - rejected as unnecessary duplication for one extra `go test` call with no meaningful runtime cost.
- **Run unconditionally, no `env:` block.** Unlike the `v1` step, this step needs no secrets, so it gets none - keeping the blast radius of any future secret rotation limited to the step that actually needs them.
- **Widen `paths` by adding `login/partner/**`, not by removing `!**_test.go`.** Adding `login/partner/**` as one more OR'd pattern (alongside the existing `v1/**`) restores triggering on `login/partner/*_test.go`-only changes without reopening the broader `!**_test.go` exclusion for every other package (which would make the workflow re-trigger on every `_test.go`-only edit repo-wide - unrelated to this change's scope).
- **Order: after the `v1` step, before the artifact upload - with an explicit `if: ${{ !cancelled() }}`.** `v1`'s tests are slower (drive a real headless Chromium against a live account) and more likely to fail for unrelated (live-account) reasons. Listing `login/partner` after `v1` keeps the diff to `test.yml` minimal and localized, but ordering alone is not cosmetic: GitHub Actions gives every step an implicit `if: success()` when it has no `if:` of its own, so without an explicit condition, a `v1` failure (which, per the tracked live-account blocker, is the common case today) would silently skip `Test login/partner` entirely - the opposite of this change's goal. `if: ${{ !cancelled() }}` makes it run whenever the job wasn't cancelled, regardless of `v1`'s outcome (caught in review; see PR #30).
- **Path filters also cover `.github/workflows/test.yml` itself.** Without it, a change that only edits this workflow file (such as this one) would match none of the existing patterns and never get validated by its own CI run (caught in review; see PR #30).
- **The live spec lives in the normal `login/partner` package, gated by `Skip(...)` on a missing env var - not a build tag.** This matches the existing repo convention (`login/chrome/chrome_test.go` already `Skip`s on missing `DIGIPOSTE_USERNAME`/`DIGIPOSTE_PASSWORD`) rather than introducing a new `-tags live` mechanism. It also means the spec still compiles and gets linted every time, just not exercised for real without the secret.
- **Select the live spec for the scheduled run with a Ginkgo label, not a separate `go test -run` pattern.** `go test -run` only filters at the level of the single `TestXxx` entrypoint each package's `suite_test.go` defines for Ginkgo, so it cannot isolate one spec inside the suite. Tagging the spec `ginkgo.Label("live")` and passing `--ginkgo.label-filter=live` (via `-args`) to the scheduled step's `go test` invocation runs only that spec, without needing a second Go test binary or file layout change.
- **Sandbox `client_id`/`client_secret` are hardcoded test constants, not secrets.** They're already public in Digiposte's own developer docs (`okapi` / `6mstVSvc38wq`); only the Okapi key (`DIGIPOSTE_OKAPI_TOKEN`) is a real credential and becomes a GitHub Actions secret. Alternative considered: making the client credentials configurable via env too - rejected as unnecessary indirection for values that aren't sensitive.
- **Gate the new scheduled step with `if: github.event_name == 'schedule'` on the same job, not a separate `schedule`-only workflow file.** Keeps the live check next to the rest of `login/partner`'s CI story in one file; the existing `tests` job already reuses its checkout/setup-go for the ordinary `login/partner` step, and the scheduled step benefits from the same reuse.

## Risks / Trade-offs

- [The screenshot-upload `Get debug screenshots` step only triggers `if: failure()` and uploads `**/*.png` - `login/partner`'s tests never produce screenshots] → No mitigation needed: it is harmless for that step to run (and find nothing to upload) on a `login/partner`-only failure.
- [Widening `paths` slightly increases how often the workflow runs] → Bounded: `login/partner/**` is a small, stable package; this is the same trade-off already accepted for `v1/**`.
- [A scheduled failure of the live spec (sandbox down, rate-limited, or Digiposte changed something) will fail the whole `tests` job on `main`, same as `v1`'s existing scheduled live-account failures already do] → Accepted, matching the precedent already set by `v1`; not a new failure mode, just one more source of the same kind of noise already tolerated on the weekly schedule.
- [The sandbox `client_id`/`client_secret` being public means anyone with a valid Okapi key for their own space could reuse them] → No mitigation needed here: that's true regardless of this CI change, and is Digiposte's sandbox design, not something this repo controls.
