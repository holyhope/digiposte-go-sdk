# digiposte-go-sdk

This repository contains the Go SDK for the [Digiposte](https://digiposte.fr) API.

![Continuous Integration](https://github.com/holyhope/digiposte-go-sdk/actions/workflows/test.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/holyhope/digiposte-go-sdk.svg)](https://pkg.go.dev/github.com/holyhope/digiposte-go-sdk)

It is a work in progress, and all the API endpoints are not implemented yet.

## Authentication

The sdk delegates the authentication to the http client. So it must be configured to add the authentication headers to the requests.

Otherwise, the [`login`](login/) package provides a simple way to authenticate and get the access token but it uses chromium to simulate a browser and is not recommended for production.
