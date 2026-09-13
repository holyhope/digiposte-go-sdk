## 1. Workflow trigger

- [x] 1.1 Add `login/partner/**` and `.github/workflows/test.yml` as additional `paths` entries to both the `push` and `pull_request` triggers in `.github/workflows/test.yml`, and verify (by re-reading the file) that a hypothetical change touching only `login/partner/okapi_test.go`, or only `test.yml` itself, would now match at least one `paths` pattern (neither did before: `!**_test.go` excludes the former and neither matched `v1/**`; the workflow file itself matched nothing at all - caught in review, see PR #30)

## 2. Test step

- [x] 2.1 Add a `Test login/partner` step to the `tests` job in `.github/workflows/test.yml`, running `go test -v ./login/partner/... -ginkgo.v` with no `env:` block, placed after the existing `Test` (`v1`) step with `if: ${{ !cancelled() }}` so a `v1` failure doesn't implicitly skip it (caught in review, see PR #30)
- [x] 2.2 Verify locally with `GOWORK=off go test -v ./login/partner/... -ginkgo.v` that the exact command added to the workflow passes

## 3. Live sandbox spec

- [x] 3.1 Add a new Ginkgo spec in `login/partner` (e.g. `login/partner/live_test.go`, `package partner_test`) tagged `ginkgo.Label("live")`, that calls `Skip("missing DIGIPOSTE_OKAPI_TOKEN")` when that env var is unset, and otherwise exercises `NewClientCredentialsSource` against `Endpoint{TokenURL: "https://api.laposte.fr/digiposte/v3/token"}` with `ClientID: "okapi"`, `ClientSecret: "6mstVSvc38wq"`, and `OkapiKey` from the env var, asserting the returned token is valid
- [x] 3.2 Verify locally with `DIGIPOSTE_OKAPI_TOKEN` unset that `GOWORK=off go test -v ./login/partner/... -ginkgo.v` still passes with the new spec reported as skipped
- [x] 3.3 Verify locally with a real `DIGIPOSTE_OKAPI_TOKEN` set that `GOWORK=off go test -v ./login/partner/... -ginkgo.v --args --ginkgo.label-filter=live` runs and passes only that spec

## 4. Scheduled CI step

- [x] 4.1 Add a `Test login/partner (live sandbox)` step to the `tests` job in `.github/workflows/test.yml`, gated `if: ${{ !cancelled() && github.event_name == 'schedule' }}`, running `go test -v ./login/partner/... -ginkgo.v --args --ginkgo.label-filter=live` with `env: DIGIPOSTE_OKAPI_TOKEN: ${{ secrets.DIGIPOSTE_OKAPI_TOKEN }}`
- [ ] 4.2 Document (in the PR description or a repo README note) that a maintainer must add the `DIGIPOSTE_OKAPI_TOKEN` repository secret for this step to actually run instead of no-op skipping

## 5. Verification

- [x] 5.1 Run `actionlint .github/workflows/test.yml` (or `yamllint -f parsable .github/workflows/test.yml` per the file's own header comment, whichever is available) and verify no new errors are introduced by the edit
- [ ] 5.2 Push the change and verify in the resulting GitHub Actions run (push/PR event) that the new `Test login/partner` step appears, passes, and reports the live spec as skipped; verify the existing `Test` (`v1`) step and `lint` job are unaffected
- [ ] 5.3 After the `DIGIPOSTE_OKAPI_TOKEN` secret is added, manually trigger or wait for the weekly `schedule` run and verify the `Test login/partner (live sandbox)` step appears and passes
