## Why

`login/partner`'s unit test suite (`go test ./login/partner/...`) never runs in CI: the `test` workflow's only test step is `go test -v ./v1/... -ginkgo.v`. The package currently only gets compile/lint coverage from the `lint` job (`golangci-lint run ./...`), so a change that builds and lints cleanly but breaks a `login/partner` test (for example, a wrong `AuthStyle` default or a broken PKCE `code_verifier`/`code_challenge` pairing) can merge to `main` undetected. Unlike `v1`'s tests, `login/partner`'s tests are self-contained (`httptest.Server` stubs only) and need no live Digiposte account or secrets, so closing this gap for every push/PR is low-cost. Separately, `login/partner`'s code has so far only been checked manually against the real Partner API sandbox (`https://api.laposte.fr/digiposte/v3/token`) - a scheduled, automated live check would catch a Digiposte-side API change (URL, response shape, grant behavior) that no unit test against a local stub could ever catch.

## What Changes

- Add a `go test ./login/partner/...` step to the `test` workflow job, run unconditionally (not gated on the live-account secrets `v1`'s tests need).
- Widen the workflow's `push`/`pull_request` path filters to explicitly include `login/partner/**`, so a change that only touches `login/partner/*_test.go` (currently excluded by the existing `!**_test.go` filter, since it lives outside `v1/**`) still triggers CI.
- Add one Ginkgo spec that exercises `NewClientCredentialsSource` against the real Digiposte Partner API sandbox (`https://api.laposte.fr/digiposte/v3/token`, confirmed to be the same base URL used for both sandbox and production), using the sandbox's publicly documented `client_id`/`client_secret` (`okapi`/`6mstVSvc38wq`) plus a real Okapi key from a new `DIGIPOSTE_OKAPI_TOKEN` secret. It self-skips (like `login/chrome`'s live specs already do for missing `DIGIPOSTE_USERNAME`) when that secret isn't set, so it's harmless in the unconditional step above.
- Add a second, weekly-`schedule`-only workflow step that supplies `DIGIPOSTE_OKAPI_TOKEN` and runs only that spec (via a Ginkgo label filter), so the live check runs on a low, predictable cadence instead of every push/PR.
- No change to what `login/partner` does at runtime; this is CI configuration only.

## Capabilities

No capability specs change; this is a CI/tooling-only change (see `skip_specs: true` in `.openspec.yaml`).

### New Capabilities

(none)

### Modified Capabilities

(none)

## Impact

- `.github/workflows/test.yml`: new unconditional test step for `login/partner`, widened `paths` filters, and a new weekly-`schedule`-only step for the live sandbox spec.
- `login/partner`: one new Ginkgo spec file, gated on a new `DIGIPOSTE_OKAPI_TOKEN` env var (Ginkgo `Skip` when absent), labeled so CI can select it independently of the rest of the suite.
- New GitHub Actions secret `DIGIPOSTE_OKAPI_TOKEN` (repo owner must add it; the sandbox `client_id`/`client_secret` themselves are public, not secret).
- No production code, package API, or dependency changes.
- Scope is deliberately limited to `login/partner`; the same CI gap exists for `login/chrome`, `login/oauth`, and `login/persistent` but is out of scope for this change (per explicit user decision).
