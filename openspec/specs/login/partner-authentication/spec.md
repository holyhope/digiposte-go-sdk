## Purpose

Lets a caller authenticate as a registered Digiposte Partenaire against the official Partner API v3, independently of and without interfering with the self-vault login methods used elsewhere in this SDK.

## Requirements

### Requirement: Client credentials authentication
The system SHALL provide a way to obtain a Partner API OAuth2 access token using the `client_credentials` grant, authenticating with a partner's `client_id`/`client_secret` and requiring no end-user interaction.

#### Scenario: Successful client credentials exchange
- **WHEN** a caller requests a token via the client credentials flow with valid `client_id`, `client_secret`, and Okapi key
- **THEN** the system returns a valid, non-expired OAuth2 access token usable as a Bearer token against the Partner API

#### Scenario: Invalid client credentials
- **WHEN** a caller requests a token via the client credentials flow with an incorrect `client_id` or `client_secret`
- **THEN** the system returns an error and no token, without retrying automatically

### Requirement: Authorization code with PKCE authentication
The system SHALL provide a way to obtain a Partner API OAuth2 access token using the `authorization_code` grant with PKCE, for flows that require a specific Digiposte end user's explicit consent.

#### Scenario: Building the authorization request
- **WHEN** a caller starts the authorization code flow with a `client_id` and `redirect_uri`
- **THEN** the system produces an authorization URL containing the `client_id`, `redirect_uri`, a `state` value, and a PKCE `code_challenge` derived from a generated `code_verifier`, and makes the matching `code_verifier` available to the caller for the later token exchange

#### Scenario: Successful authorization code exchange
- **WHEN** a caller exchanges a valid `code` returned by Digiposte, together with the original `code_verifier` and `redirect_uri`, for a token
- **THEN** the system returns a valid OAuth2 access token (and refresh token, when Digiposte issues one) usable as a Bearer token against the Partner API

#### Scenario: Invalid or expired authorization code
- **WHEN** a caller exchanges a `code` that is invalid, expired, or already used
- **THEN** the system returns an error and no token

#### Scenario: State mismatch
- **WHEN** a caller supplies a `state` value on the callback that does not match the `state` used to build the authorization URL
- **THEN** the system SHALL treat this as an error and SHALL NOT proceed with the token exchange

### Requirement: Okapi key on every Partner API request
Every request made by either authentication flow to the Partner API SHALL include a caller-supplied `X-Okapi-Key` header, distinguishing it from the self-vault login flows, which use no such header.

#### Scenario: Missing Okapi key
- **WHEN** a caller attempts either authentication flow without configuring an Okapi key
- **THEN** the system returns a configuration error before making any request to the Partner API

### Requirement: Environment-selectable endpoints
The system SHALL support directing either authentication flow at either Digiposte's sandbox or production Partner API environment, each with its own base URLs, without code changes beyond configuration.

#### Scenario: Sandbox selected
- **WHEN** a caller configures either flow for the sandbox environment
- **THEN** the system sends authorization and token requests to the sandbox Partner API endpoints

#### Scenario: Production selected
- **WHEN** a caller configures either flow for the production environment
- **THEN** the system sends authorization and token requests to the production Partner API endpoints

### Requirement: Independence from self-vault authentication
Partner authentication SHALL operate independently of the self-vault `login.Method`/`login/chrome`/`login/persistent` flows: obtaining, storing, or using a Partner API token SHALL NOT require, consume, or mutate any self-vault session, cookie, or credential, and vice versa.

#### Scenario: Partner token obtained without a self-vault session
- **WHEN** a caller obtains a Partner API token via either flow without ever having established a self-vault session
- **THEN** the system succeeds, with no dependency on `login.Credentials` or any self-vault cookie/token
