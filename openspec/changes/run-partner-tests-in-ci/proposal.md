## Why

`login/partner`'s unit test suite (`go test ./login/partner/...`) never runs in CI: the `test` workflow's only test step is `go test -v ./v1/... -ginkgo.v`. The package currently only gets compile/lint coverage from the `lint` job (`golangci-lint run ./...`), so a change that builds and lints cleanly but breaks a `login/partner` test (for example, a wrong `AuthStyle` default or a broken PKCE `code_verifier`/`code_challenge` pairing) can merge to `main` undetected. Unlike `v1`'s tests, `login/partner`'s tests are self-contained (`httptest.Server` stubs only) and need no live Digiposte account or secrets, so this is a low-cost gap to close.

## What Changes

- Add a `go test ./login/partner/...` step to the `test` workflow job, run unconditionally (not gated on the live-account secrets `v1`'s tests need).
- Widen the workflow's `push`/`pull_request` path filters to explicitly include `login/partner/**`, so a change that only touches `login/partner/*_test.go` (currently excluded by the existing `!**_test.go` filter, since it lives outside `v1/**`) still triggers CI.
- No change to what `login/partner` does at runtime; this is CI configuration only.

## Capabilities

No capability specs change; this is a CI/tooling-only change (see `skip_specs: true` in `.openspec.yaml`).

### New Capabilities

(none)

### Modified Capabilities

(none)

## Impact

- `.github/workflows/test.yml`: new test step for `login/partner`, widened `paths` filters.
- No production code, package API, or dependency changes.
- Scope is deliberately limited to `login/partner`; the same CI gap exists for `login/chrome`, `login/oauth`, and `login/persistent` but is out of scope for this change (per explicit user decision).
