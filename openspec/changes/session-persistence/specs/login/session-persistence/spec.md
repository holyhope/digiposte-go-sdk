## Purpose

Lets an SDK caller reuse a previously established Digiposte session across process runs, so that authentication only drives an interactive (browser-based) login when no valid session is available, instead of on every call.

## ADDED Requirements

### Requirement: Session Store contract
The system SHALL define a store contract for loading and saving a session, where a session consists of an OAuth token and its associated cookies. The contract SHALL allow loading to report that no session is currently stored, distinct from an error.

#### Scenario: No session stored yet
- **WHEN** a session-resuming login is attempted and the configured store reports no stored session
- **THEN** the system proceeds as if no prior session existed, and does not treat the absence of a session as an error

#### Scenario: Store is optional
- **WHEN** an SDK caller constructs a login method without configuring any store
- **THEN** the system performs a plain interactive login on every call and does not attempt to load or save a session

### Requirement: Resume a still-valid stored session without interactive login
The system SHALL, when a configured store returns a session whose token is still valid, return that session directly without performing an interactive (browser-based) login.

#### Scenario: Valid token is resumed
- **WHEN** a session-resuming login is attempted and the store returns a token that is not expired
- **THEN** the system returns that token and its cookies without launching or driving a browser

### Requirement: Fall back to interactive login when the stored session is invalid or expired
The system SHALL perform a full interactive login when the store has no session, when the stored token is expired or invalid, or when the store returns an error while loading.

#### Scenario: Expired token falls back to interactive login
- **WHEN** a session-resuming login is attempted and the store returns a session whose token is expired
- **THEN** the system performs an interactive login instead of returning the expired token

#### Scenario: Store load failure falls back to interactive login
- **WHEN** a session-resuming login is attempted and the store returns an error while loading a session
- **THEN** the system performs an interactive login instead of failing the overall login attempt

### Requirement: Seed a previously stored session into an interactive login attempt
When falling back to an interactive login after loading a stored session (valid or not), the system SHALL make the stored session's cookies available to that interactive login attempt, so that a still-authenticated site session can be recognized without re-entering credentials.

#### Scenario: Interactive login recognizes a still-valid site session from seeded cookies
- **WHEN** an interactive login is attempted with cookies seeded from a previously stored session, and those cookies still represent an authenticated site session
- **THEN** the interactive login completes successfully without presenting credential, one-time-passcode, or trusted-device screens to the caller

#### Scenario: Interactive login falls back to the full flow when seeded cookies are no longer valid
- **WHEN** an interactive login is attempted with seeded cookies that no longer represent an authenticated site session
- **THEN** the interactive login presents the normal credential, one-time-passcode, and trusted-device screens as if no cookies had been seeded

### Requirement: Persist the outcome of every successful login
The system SHALL save the resulting token and cookies to the configured store after every successful login, whether the session was resumed or obtained interactively. A failure to save SHALL be reported to the caller without causing an otherwise-successful login to fail.

#### Scenario: Successful interactive login is saved
- **WHEN** an interactive login completes successfully and a store is configured
- **THEN** the resulting token and cookies are saved to the store before the login call returns successfully

#### Scenario: Save failure does not fail the login
- **WHEN** an interactive or resumed login completes successfully but the configured store returns an error while saving
- **THEN** the login call still returns the obtained token and cookies successfully, and the save error is reported separately rather than as a login failure

### Requirement: No implicit persistence to disk
The system SHALL NOT persist a session to any storage medium unless the caller has explicitly configured a store. The system SHALL NOT ship a default store implementation that writes session data to disk automatically.

#### Scenario: No store configured means no storage side effect
- **WHEN** an SDK caller uses the session-resuming login method without configuring a store
- **THEN** no session data is written to disk, a keychain, or any other persistent medium by the system
