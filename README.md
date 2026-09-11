# digiposte-go-sdk

![Continuous Integration](https://github.com/holyhope/digiposte-go-sdk/actions/workflows/test.yml/badge.svg)
[![Go References](https://pkg.go.dev/badge/github.com/holyhope/digiposte-go-sdk.svg)](https://pkg.go.dev/github.com/holyhope/digiposte-go-sdk)

This repository contains the Go SDK for the [Digiposte](https://digiposte.fr) API.

It is a work in progress, and all the API endpoints are not implemented yet.

Last run succeeded on `2024-07-28`.

## Authentication

The sdk delegates the authentication to the http client. So it must be configured to add the authentication headers to the requests.

Otherwise, the [`login`](login/) package provides a simple way to authenticate and get the access token but it uses chromium to simulate a browser and is not recommended for production.

Driving that browser-based login on every run is fragile and can trigger rate limiting. The [`login/persistent`](login/persistent/) package lets a caller reuse a previously obtained session instead: it defines a `Store` contract for loading/saving a session (an OAuth token and its cookies) and a `login.Method` decorator that resumes a still-valid stored session with no browser at all, or falls back to an interactive login otherwise, seeding any stored cookies into it so an already-authenticated site session is recognized without re-entering credentials. `login/persistent` ships **no default `Store` implementation** - a stored session is equivalent to live account credentials, so callers must provide and protect their own storage (see the `Store` doc comment).
