# digiposte-go-sdk

![Continuous Integration](https://github.com/holyhope/digiposte-go-sdk/actions/workflows/test.yml/badge.svg)
[![Go References](https://pkg.go.dev/badge/github.com/holyhope/digiposte-go-sdk.svg)](https://pkg.go.dev/github.com/holyhope/digiposte-go-sdk)

Go SDK for the [Digiposte](https://digiposte.fr) API.

It is a work in progress, and not all API endpoints are implemented yet.

## Installation

```sh
go get github.com/holyhope/digiposte-go-sdk
```

## Quick start

```go
import (
	"context"
	"net/http"
	"os"

	"github.com/holyhope/digiposte-go-sdk/login"
	digiposte "github.com/holyhope/digiposte-go-sdk/v1"
)

client, err := digiposte.NewAuthenticatedClient(context.Background(), http.DefaultClient, &digiposte.Config{
	Credentials: &login.Credentials{
		Username:  os.Getenv("DIGIPOSTE_USERNAME"),
		Password:  os.Getenv("DIGIPOSTE_PASSWORD"),
		OTPSecret: os.Getenv("DIGIPOSTE_OTP_SECRET"),
	},
})
if err != nil {
	// handle error
}

// client.CreateFolder(...), client.CreateDocument(...), client.DocumentContent(...), client.Trash(...), ...
```

`NewAuthenticatedClient` drives an interactive, browser-based login (see [Authentication](#authentication)) the first time it needs a token, then reuses it for subsequent requests. See [`v1/example_test.go`](v1/example_test.go)'s `Example` for a full runnable walkthrough (create a folder, upload a document, read it back, then clean up).

## Packages

| Package | Purpose |
| --- | --- |
| `v1` (repository root import path) | The Digiposte API client itself: folders, documents, sharing, trash. |
| [`login`](login/) | Shared login types used by every method below: `Credentials`, the `Method` interface. |
| [`login/chrome`](login/chrome/) | Interactive, browser-driven login. Drives a real (headless) Chromium instance through Digiposte's credentials, one-time-passcode, and trusted-device screens. It is the only way to obtain a session today, and is slow and rate-limit-sensitive if run on every process start. |
| [`login/oauth`](login/oauth/) | Adapts a `login.Method` into an `oauth2.TokenSource`, so it composes with `oauth2.ReuseTokenSource` and `oauth2.Transport` like any other token source. |
| [`login/persistent`](login/persistent/) | Decorates a `login.Method` so a previously obtained session is resumed instead of driving the interactive login again on every run - see below. |

## Authentication

The SDK does not manage authentication for you by default - `digiposte.NewClient(client)` takes a plain `*http.Client` and expects it to already carry whatever authentication headers/cookies are needed for it to work. `NewAuthenticatedClient` builds one of those clients for you, using one of the following login strategies.

### Interactive login (`login/chrome`)

The [`login`](login/) package provides a simple way to authenticate and get an access token, but it uses a headless Chromium browser to simulate a human logging in, and is not recommended for production use: it is slow, fragile to Digiposte's page markup changing, and can trigger rate limiting on the account if run too often.

### Reusing a session across runs (`login/persistent`)

Driving that browser-based login on every run is fragile and can trigger rate limiting. The [`login/persistent`](login/persistent/) package lets a caller reuse a previously obtained session instead: it defines a `Store` contract for loading/saving a session (an OAuth token and its cookies) and a `login.Method` decorator that resumes a still-valid stored session with no browser at all, or falls back to an interactive login otherwise, seeding any stored cookies into it so an already-authenticated site session is recognized without re-entering credentials.

`login/persistent` ships **no default `Store` implementation** - a stored session is equivalent to live account credentials, so callers must provide and protect their own storage (see the `Store` doc comment). See [`login/persistent/example_test.go`](login/persistent/example_test.go)'s `ExampleNew` for the full composition pattern, and `v1.Config`'s `PreviousSession`/`SessionListener` fields for wiring the same persistence directly into the `v1` client.

## Status

See the CI badge above for the latest build status, and [`openspec/`](openspec/) for in-progress design proposals if you want to see where the project is headed.

## License

Mozilla Public License 2.0 - see [LICENSE](LICENSE).
