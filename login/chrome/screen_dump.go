package chrome

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// networkCapturer records CDP network responses to disk for as long as it
// is listening. It is an interface - rather than Screens.run calling
// chromedp/cdproto directly - so a unit test can inject a fake capturer and
// verify Screens.run's wiring (called only when screenDumpDir is set)
// without a live browser target. See screens_internal_test.go.
type networkCapturer interface {
	// capture starts recording every network response observed on ctx's
	// browser target into dir, organized by real hostname/path, for as
	// long as ctx stays open. It returns once recording has started (or
	// failed to start); it does not block for the lifetime of ctx.
	capture(ctx context.Context, dir string) error
}

// cdpNetworkCapturer is the production networkCapturer, backed by chromedp
// and the CDP Network domain. It is only ever used when screenDumpDir is
// set, which in turn is only ever set by the test-only WithScreenDumpDir
// helper in export_test.go.
type cdpNetworkCapturer struct{}

func (cdpNetworkCapturer) capture(ctx context.Context, dir string) error {
	err := chromedp.Run(ctx, network.Enable())
	if err != nil {
		return fmt.Errorf("enable network domain: %w", err)
	}

	requestURLs := map[network.RequestID]string{}

	// writeMu serializes the exists-check-then-write in
	// writeCapturedResponse (via dumpResponseBody) across the concurrent
	// per-request goroutines started below, so two responses landing on
	// the same destination path race-free choose distinct collision
	// suffixes rather than one clobbering the other's exists-check.
	var writeMu sync.Mutex

	chromedp.ListenTarget(ctx, func(event any) {
		handleNetworkEvent(ctx, event, requestURLs, dir, &writeMu)
	})

	return nil
}

func handleNetworkEvent(
	ctx context.Context,
	event any,
	requestURLs map[network.RequestID]string,
	dir string,
	writeMu *sync.Mutex,
) {
	switch event := event.(type) {
	case *network.EventResponseReceived:
		if event.Response != nil {
			requestURLs[event.RequestID] = event.Response.URL
		}

	case *network.EventLoadingFinished:
		rawURL, ok := requestURLs[event.RequestID]
		if !ok {
			return
		}

		delete(requestURLs, event.RequestID)

		// GetResponseBody makes a further CDP round-trip; run it off the
		// event-dispatch goroutine so it never blocks delivery of other
		// events, as chromedp's own examples do.
		go dumpResponseBody(ctx, event.RequestID, rawURL, dir, writeMu)
	}
}

func dumpResponseBody(ctx context.Context, requestID network.RequestID, rawURL, dir string, writeMu *sync.Mutex) {
	if ctx.Err() != nil {
		return
	}

	// network.GetResponseBody must be run through chromedp.Run, not
	// called directly on ctx: cdproto commands read their executor from a
	// context value that only chromedp.Run injects before calling
	// action.Do(ctx). Calling .Do(ctx) directly fails with "invalid
	// context" every time - see the chromedp README ("Executing an action
	// without Run results in 'invalid context'").
	var body []byte

	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var getErr error

		body, getErr = network.GetResponseBody(requestID).Do(ctx)
		if getErr != nil {
			return fmt.Errorf("get response body: %w", getErr)
		}

		return nil
	}))
	if err != nil {
		errorLogger(ctx).Printf("Failed to get response body for %q: %v\n", rawURL, err)

		return
	}

	writeMu.Lock()
	err = writeCapturedResponse(dir, rawURL, body)
	writeMu.Unlock()

	if err != nil {
		errorLogger(ctx).Printf("Failed to write captured response for %q: %v\n", rawURL, err)

		return
	}

	infoLogger(ctx).Printf("Captured %q\n", rawURL)
}

var errCaptureNoHostname = errors.New("captured response has no hostname")

const (
	dirPerm  = 0o755
	filePerm = 0o644
)

// writeCapturedResponse writes body to <dir>/<hostname>/<path>, mirroring
// rawURL's real hostname and URL path (the query string, if any, is
// dropped: fixtures are looked up by hostname+path only, see
// mock_server_test.go). A path that is empty or ends in "/" is stored as
// "index.html" within that directory, matching a conventional static-site
// mirror; any ".." path segment is stripped so a captured response can
// never be written outside dir.
//
// Multi-step flows (for example Keycloak serving the credentials, OTP, and
// trusted-device screens from the exact same URL) would otherwise silently
// overwrite each earlier capture at that path. To preserve every distinct
// response instead, a path that already exists on disk is not overwritten;
// a numeric suffix is appended before the extension (authenticate,
// authenticate-2, authenticate-3, ...) until a free path is found. A
// maintainer curating fixtures (see tasks.md 2.1) is expected to inspect
// and rename these numbered files to descriptive ones.
func writeCapturedResponse(dir, rawURL string, body []byte) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse captured url %q: %w", rawURL, err)
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("%q: %w", rawURL, errCaptureNoHostname)
	}

	relPath := parsed.Path
	if relPath == "" || relPath[len(relPath)-1] == '/' {
		relPath += "index.html"
	}

	// Clean relative to a synthetic root so any ".." segment collapses
	// instead of escaping dir.
	cleanPath := filepath.Clean(filepath.Join(string(filepath.Separator), relPath))

	dest := filepath.Join(dir, hostname, cleanPath)

	err = os.MkdirAll(filepath.Dir(dest), dirPerm)
	if err != nil {
		return fmt.Errorf("mkdir %q: %w", filepath.Dir(dest), err)
	}

	dest, err = nextAvailablePath(dest)
	if err != nil {
		return fmt.Errorf("find free path for %q: %w", dest, err)
	}

	err = os.WriteFile(dest, body, filePerm)
	if err != nil {
		return fmt.Errorf("write %q: %w", dest, err)
	}

	return nil
}

// nextAvailablePath returns dest unchanged if nothing exists there yet, or
// the first of dest, dest-2, dest-3, ... (extension preserved) that does
// not already exist.
func nextAvailablePath(dest string) (string, error) {
	_, err := os.Stat(dest)

	switch {
	case errors.Is(err, os.ErrNotExist):
		return dest, nil
	case err != nil:
		return "", fmt.Errorf("stat %q: %w", dest, err)
	}

	ext := filepath.Ext(dest)
	base := strings.TrimSuffix(dest, ext)

	for suffix := 2; ; suffix++ {
		candidate := fmt.Sprintf("%s-%d%s", base, suffix, ext)

		_, err := os.Stat(candidate)

		switch {
		case errors.Is(err, os.ErrNotExist):
			return candidate, nil
		case err != nil:
			return "", fmt.Errorf("stat %q: %w", candidate, err)
		}
	}
}
