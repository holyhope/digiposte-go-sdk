## Context

`.github/workflows/test.yml` has one job (`tests`) with one test step, `go test -v ./v1/... -ginkgo.v`, and `push`/`pull_request` triggers gated on `paths: ['**.go', '!**_test.go', 'v1/**']`. See `proposal.md` - Why for the coverage gap this leaves for `login/partner`.

## Goals / Non-Goals

**Goals:**

- Make a `login/partner` test failure fail CI, the same way a `v1` test failure does today.
- Run without any of the `DIGIPOSTE_*` secrets/vars the `v1` step needs - `login/partner`'s tests only use `httptest.Server` stubs.
- Trigger on a `login/partner`-only test-file change, which the current path filter drops.

**Non-Goals:**

- Closing the same CI gap for `login/chrome`, `login/oauth`, or `login/persistent` (explicitly out of scope per proposal.md - Impact).
- Changing anything about how `v1`'s tests run, or the weekly `schedule` trigger.
- Adding a separate GitHub Actions job. One extra step in the existing `tests` job is enough here; a second job would duplicate the checkout/setup-go steps for a single `go test` invocation.

## Decisions

- **New step in the existing job, not a new job.** The existing job already has Go set up; adding `- name: Test login/partner` \ `run: go test -v ./login/partner/... -ginkgo.v` after the current `Test` step reuses that setup and keeps a single place to look for test results. Alternative considered: a separate `partner-tests` job running in parallel - rejected as unnecessary duplication for one extra `go test` call with no meaningful runtime cost.
- **Run unconditionally, no `env:` block.** Unlike the `v1` step, this step needs no secrets, so it gets none - keeping the blast radius of any future secret rotation limited to the step that actually needs them.
- **Widen `paths` by adding `login/partner/**`, not by removing `!**_test.go`.** Adding `login/partner/**` as one more OR'd pattern (alongside the existing `v1/**`) restores triggering on `login/partner/*_test.go`-only changes without reopening the broader `!**_test.go` exclusion for every other package (which would make the workflow re-trigger on every `_test.go`-only edit repo-wide - unrelated to this change's scope).
- **Order: after the `v1` step, before the artifact upload.** `v1`'s tests are slower (drive a real headless Chromium against a live account) and more likely to fail for unrelated (live-account) reasons; running the fast, self-contained `login/partner` step second still fails the job either way, so ordering is cosmetic - listing it after `v1` keeps the diff to `test.yml` minimal and localized.

## Risks / Trade-offs

- [The screenshot-upload `Get debug screenshots` step only triggers `if: failure()` and uploads `**/*.png` - `login/partner`'s tests never produce screenshots] → No mitigation needed: it is harmless for that step to run (and find nothing to upload) on a `login/partner`-only failure.
- [Widening `paths` slightly increases how often the workflow runs] → Bounded: `login/partner/**` is a small, stable package; this is the same trade-off already accepted for `v1/**`.
