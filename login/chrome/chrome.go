package chrome

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login"
)

type WithLocationError struct {
	Err      error
	Location string
}

func (e *WithLocationError) Error() string {
	return fmt.Sprintf("%v at %q", e.Err, e.Location)
}

func (c *chromeLogin) WrapError(ctx context.Context, errPtr *error) {
	if *errPtr == nil {
		return
	}

	var currentLocation string

	err := chromedp.Run(ctx,
		chromedp.Location(&currentLocation),
	)
	if err != nil {
		errorLogger(ctx).Printf("Failed to get current location: %v\n", err)
	} else {
		*errPtr = &WithLocationError{
			Err:      *errPtr,
			Location: currentLocation,
		}
	}

	*errPtr = c.ScreenshotIfNeeded(ctx, *errPtr)
}

func (c *chromeLogin) login(
	parentCtx, independentChromeCtx context.Context,
	creds *login.Credentials,
) (_ *oauth2.Token, _ []*http.Cookie, finalErr error) { //nolint:nonamedreturns
	if c.timeout > 0 {
		ctx, cancel := context.WithTimeout(parentCtx, c.timeout)
		defer cancel()

		parentCtx = ctx
	}

	ctx, cancel := WithCancelOnClose(independentChromeCtx, parentCtx.Done())
	defer cancel()

	defer c.WrapError(independentChromeCtx, &finalErr)

	err := resolve(ctx, &firstScreen{
		URL:     c.url,
		Cookies: c.cookies,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("first screen: %w", err)
	}

	infoLogger(ctx).Printf("Page %q loaded\n", c.url)

	return c.resolveLogin(ctx, creds)
}

func WithCancelOnClose(ctx context.Context, done <-chan struct{}) (context.Context, context.CancelFunc) {
	attachedChromeCtx, cancel := context.WithCancel(ctx)

	go func(independentChromeCtx context.Context, done <-chan struct{}, cancel context.CancelFunc) {
		select {
		case <-independentChromeCtx.Done(): // do nothing
		case <-done:
			cancel()
		}
	}(ctx, done, cancel)

	return attachedChromeCtx, cancel
}

func (c *chromeLogin) resolveLogin(
	ctx context.Context,
	creds *login.Credentials,
) (*oauth2.Token, []*http.Cookie, error) {
	finalScreen := &finalScreen{
		Token:   nil,
		Cookies: nil,
	}

	screens := Screens{
		screens: []Screen{
			&privacyScreen{
				AcceptCookies: false,
			},
			&credentialsScreen{
				Username:  creds.Username,
				Password:  creds.Password,
				submitted: atomic.Bool{},
			},
			&otpScreen{
				Secret: creds.OTPSecret,
			},
			&trustedDeviceScreen{},
			finalScreen,
		},
		refreshFrequency: c.refreshFrequency,
		screenTimeout:    c.screenTimeout,
		resolver:         chromedpResolver{},
		succeeded:        atomic.Bool{},
	}

	go screens.Resolve(ctx)

	ticker := time.NewTicker(c.refreshFrequency)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("context done: %w", ctx.Err())

		case <-ticker.C:
			if finalScreen.Token != nil {
				screens.succeeded.Store(true)

				return finalScreen.Token, finalScreen.Cookies, nil
			}
		}
	}
}

type chromeLogin struct {
	url string

	cookies []*http.Cookie

	screenShortOnError bool
	refreshFrequency   time.Duration
	screenTimeout      time.Duration
	timeout            time.Duration

	infoLogger  *log.Logger
	errorLogger *log.Logger

	binaryPath string
}

type HTTPError struct {
	Status     int64
	StatusText string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error %d: %s", e.Status, e.StatusText)
}
