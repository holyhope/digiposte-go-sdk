package digiposte_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"golang.org/x/time/rate"

	digipoauth "github.com/holyhope/digiposte-go-sdk/login"
	"github.com/holyhope/digiposte-go-sdk/login/chrome"
	"github.com/holyhope/digiposte-go-sdk/v1"
)

func DigiposteClient(ctx context.Context) (*digiposte.Client, error) {
	digiposteClientLock.Lock()
	defer digiposteClientLock.Unlock()

	if digiposteClient != nil {
		return digiposteClient, nil
	}

	// or use digiposte.DefaultDocumentURL
	// Reduce the test duration
	client, err := newDigiposteClient(ctx)
	if err != nil {
		return nil, err
	}

	digiposteClient = client

	return digiposteClient, nil
}

func newDigiposteClient(ctx context.Context) (*digiposte.Client, error) {
	path, err := getChrome(ctx)
	if err != nil {
		return nil, fmt.Errorf("get chrome: %w", err)
	}

	documentURL := os.Getenv("DIGIPOSTE_URL")

	chromeMethod, err := chrome.New(
		chrome.WithURL(documentURL),
		chrome.WithRefreshFrequency(500*time.Millisecond),
		chrome.WithScreenShortOnError(),
		chrome.WithTimeout(3*time.Minute),
		chrome.WithBinary(path),
	)
	if err != nil {
		return nil, fmt.Errorf("new chrome: %w", err)
	}

	client, err := digiposte.NewAuthenticatedClient(ctx, &http.Client{
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
		Transport: &rateLimitedTransport{
			RoundTripper: http.DefaultTransport,
			rateLimiter:  rate.NewLimiter(rate.Every(1*time.Second), 5),
		},
	}, &digiposte.Config{
		APIURL:      os.Getenv("DIGIPOSTE_API"),
		DocumentURL: documentURL,
		LoginMethod: chromeMethod,
		Credentials: &digipoauth.Credentials{
			Username:  os.Getenv("DIGIPOSTE_USERNAME"),
			Password:  os.Getenv("DIGIPOSTE_PASSWORD"),
			OTPSecret: os.Getenv("DIGIPOSTE_OTP_SECRET"),
		},
	})
	if err != nil {
		screenshot, ok := chrome.GetScreenShot(err)
		if ok {
			if err := os.WriteFile("screenshot.png", screenshot, 0o600); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save the screenshot: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Screenshot saved to %q\n", "screenshot.png")
			}
		}

		return nil, fmt.Errorf("new client: %w", err)
	}

	return client, nil
}

func getChrome(ctx context.Context) (string, error) {
	browser := launcher.NewBrowser()

	browser.Context = ctx

	path, err := browser.Get()
	if err != nil {
		return "", fmt.Errorf("browser download: %w", err)
	}

	if err := browser.Validate(); err != nil {
		return "", fmt.Errorf("browser download validation: %w", err)
	}

	return path, nil
}

type rateLimitedTransport struct {
	http.RoundTripper
	rateLimiter *rate.Limiter
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.rateLimiter.Wait(req.Context()); err != nil {
		return nil, fmt.Errorf("rate limited: %w", err)
	}

	resp, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("round trip: %w", err)
	}

	return resp, nil
}

//nolint:gochecknoglobals
var (
	digiposteClient     *digiposte.Client
	digiposteClientLock sync.Mutex
)

var _ = ginkgo.BeforeSuite(func(ctx ginkgo.SpecContext) {
	gomega.Expect(DigiposteClient(ctx)).NotTo(gomega.BeNil())
})

func TestV1(t *testing.T) {
	t.Parallel()

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "V1 Suite")
}
