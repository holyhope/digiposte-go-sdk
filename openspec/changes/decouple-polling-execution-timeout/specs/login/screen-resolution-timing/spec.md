## Purpose

Governs how long the Chrome-driven login flow polls for a screen versus how long it is given to actually complete that screen's automation, so a fast polling cadence never truncates a slower, multi-step screen mid-action.

## ADDED Requirements

### Requirement: Independent polling frequency and screen execution timeout
The system SHALL treat "how often a screen is checked for a match" and "how long a matched screen's automation is allowed to run" as two independently configurable durations. Configuring one SHALL NOT change the other's effective value.

#### Scenario: Short polling frequency does not shorten screen execution time
- **WHEN** a caller configures a short screen-polling frequency (e.g. 500ms) without configuring the screen execution timeout
- **THEN** a matched screen's automation SHALL still be allowed to run for the full screen execution timeout (its own configured or default value), rather than being bounded by the shorter polling frequency

#### Scenario: Screen execution timeout does not change polling cadence
- **WHEN** a caller configures a longer screen execution timeout without configuring the screen-polling frequency
- **THEN** the system SHALL continue polling for a screen match at the existing polling frequency, unaffected by the new execution timeout value

### Requirement: Default execution timeout accommodates built-in multi-step screens
The system SHALL provide a default screen execution timeout that is long enough, under normal network conditions against the real Digiposte login pages, for each of the login flow's built-in multi-step screens (credentials entry, OTP entry, trusted-device confirmation) to complete its automation in a single attempt, without requiring a caller to configure anything.

#### Scenario: Multi-step screen completes within the default timeout
- **WHEN** the login flow is run with only a screen-polling frequency configured (default execution timeout, default or short polling frequency) against a reachable, responsive login page
- **THEN** the credentials, OTP, and trusted-device screens each complete their automation (submit successfully) without being interrupted by the execution timeout

### Requirement: Execution timeout is configurable independently of polling frequency
The system SHALL expose a way for a caller to configure the screen execution timeout, distinct from the existing polling-frequency configuration, so a caller with slower automation needs (e.g. a slower network) can raise it without also having to slow down polling.

#### Scenario: Caller raises the execution timeout only
- **WHEN** a caller configures a longer screen execution timeout than the default
- **THEN** a matched screen's automation is allowed to run up to that longer duration before being interrupted, regardless of the polling frequency in effect

#### Scenario: Invalid execution timeout is rejected
- **WHEN** a caller configures a non-positive screen execution timeout
- **THEN** the system SHALL reject the configuration with an error before starting the login flow, and SHALL NOT silently fall back to a default
