## 1. Workflow trigger

- [ ] 1.1 Add `login/partner/**` as an additional `paths` entry to both the `push` and `pull_request` triggers in `.github/workflows/test.yml`, and verify (by re-reading the file) that a hypothetical change touching only `login/partner/okapi_test.go` would now match at least one `paths` pattern (it doesn't today, since it's excluded by `!**_test.go` and isn't under `v1/**`)

## 2. Test step

- [ ] 2.1 Add a `Test login/partner` step to the `tests` job in `.github/workflows/test.yml`, running `go test -v ./login/partner/... -ginkgo.v` with no `env:` block, placed after the existing `Test` (`v1`) step
- [ ] 2.2 Verify locally with `GOWORK=off go test -v ./login/partner/... -ginkgo.v` that the exact command added to the workflow passes

## 3. Verification

- [ ] 3.1 Run `actionlint .github/workflows/test.yml` (or `yamllint -f parsable .github/workflows/test.yml` per the file's own header comment, whichever is available) and verify no new errors are introduced by the edit
- [ ] 3.2 Push the change and verify in the resulting GitHub Actions run that the new `Test login/partner` step appears and passes, and that the existing `Test` (`v1`) step and `lint` job are unaffected
