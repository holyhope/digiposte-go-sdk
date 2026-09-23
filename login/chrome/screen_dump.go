package chrome

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

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

	chromedp.ListenTarget(ctx, func(event any) {
		handleNetworkEvent(ctx, event, requestURLs, dir)
	})

	return nil
}

func handleNetworkEvent(ctx context.Context, event any, requestURLs map[network.RequestID]string, dir string) {
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
		go dumpResponseBody(ctx, event.RequestID, rawURL, dir)
	}
}

func dumpResponseBody(ctx context.Context, requestID network.RequestID, rawURL, dir string) {
	if ctx.Err() != nil {
		return
	}

	body, err := network.GetResponseBody(requestID).Do(ctx)
	if err != nil {
		errorLogger(ctx).Printf("Failed to get response body for %q: %v\n", rawURL, err)

		return
	}

	err = writeCapturedResponse(dir, rawURL, body)
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

	err = os.WriteFile(dest, body, filePerm)
	if err != nil {
		return fmt.Errorf("write %q: %w", dest, err)
	}

	return nil
}
