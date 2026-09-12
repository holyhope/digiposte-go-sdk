## Why

Every login method in this SDK today (`login/chrome`, decorated by `login/persistent`) drives or resumes a single, specific *end user's* interactive session against Digiposte's private self-vault surface (`secure.digiposte.fr` / `api.digiposte.fr`) - there is no way to authenticate as a registered Digiposte **Partenaire** (the [official Partner API v3](https://developer.laposte.fr/catalog-apis/digiposte@3), fronted by La Poste's OKAPI gateway at `api.laposte.fr/digiposte/v3`). The Partner API uses real OAuth2 (`client_credentials` for backend/server-to-server calls, `authorization_code` + PKCE for flows that need a specific end user's consent) and every request additionally requires an `X-Okapi-Key` header - a completely different auth model from the browser-scraped session the SDK produces today. Without this, the SDK can only ever act as "ourselves" against our own vault; it cannot act as a partner organization on behalf of (or in relation to) other Digiposte accounts.

## What Changes

- Add a new `login/partner` package implementing two independent ways to obtain a Partner API OAuth2 token, each returning a standard `*oauth2.Token`:
  - `client_credentials` grant (server-to-server, HTTP Basic `client_id:client_secret`, no end-user interaction) - via `golang.org/x/oauth2/clientcredentials`.
  - `authorization_code` + PKCE grant (redirects the end user to Digiposte for consent, then exchanges the returned `code` + `code_verifier` for a token) - via `golang.org/x/oauth2`'s existing `GenerateVerifier`/`S256ChallengeOption`/`VerifierOption` helpers.
- Every request made with either flow (token requests and, later, resource requests) must carry an `X-Okapi-Key` header; `login/partner` is responsible for attaching it consistently for both grants.
- Both flows target the Partner API's own OAuth endpoints and base URLs (sandbox: `auth.interop.digiposte.io`/`api.laposte.fr`; production: `secure.digiposte.fr/identification-plus`/`api.laposte.fr`) - entirely separate from `settings.DefaultAPIURL`/`DefaultDocumentURL`, which remain self-vault-only and untouched.
- No changes to the existing `login`, `login/chrome`, `login/oauth`, `login/persistent`, `login/noop`, or `v1` packages, and no changes to their public APIs. This change is purely additive.

**Explicitly out of scope** (left for a later change): the partner API operations this authentication unlocks - creating/checking a membership (adhésion) between a `partner_user_id` and a Digiposte account, and depositing/retrieving documents in a member's vault as a partner. This change only produces a usable OAuth2 token and the `X-Okapi-Key`-aware plumbing to call those endpoints later; it does not implement the endpoints themselves.

## Capabilities

### New Capabilities

- `login/partner-authentication`: Defines the two supported ways to obtain a Digiposte Partner API OAuth2 token (`client_credentials` and `authorization_code`+PKCE), the `X-Okapi-Key` header requirement, and how each flow's inputs/outputs are shaped - independent of and non-overlapping with the existing self-vault `login` capability.

### Modified Capabilities

(none - this change is purely additive; no existing capability's requirements change)

## Impact

- **Affected code**: new package `login/partner` (client_credentials flow, authorization_code+PKCE flow, `X-Okapi-Key` plumbing, sandbox/production endpoint configuration). No existing package is modified.
- **Dependencies**: no new third-party dependencies - `golang.org/x/oauth2/clientcredentials` and the PKCE helpers (`oauth2.GenerateVerifier`, `oauth2.S256ChallengeOption`, `oauth2.VerifierOption`) are already available via the existing `golang.org/x/oauth2` requirement.
- **Compatibility**: fully additive; no breaking changes to `login`, `login/chrome`, `login/oauth`, `login/persistent`, `login/noop`, or `v1`, and no changes to their import paths.
- **Documentation**: `README.md`'s package table and Authentication section need a new entry describing `login/partner` and how it differs from the self-vault login methods above it.
- **Not affected**: `v1` (self-vault client), `settings` (self-vault base URLs), and any existing consumer (e.g. the `rclone` Digiposte backend) - none of them import or depend on the new package.
