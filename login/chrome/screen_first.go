package chrome

import (
	"context"
	"fmt"
	"net/http"

	"github.com/chromedp/chromedp"
)

type firstScreen struct {
	URL string

	// Cookies, when non-empty, are seeded into the browser before URL is
	// navigated to, so that a still-authenticated site session is recognized
	// without driving the interactive login screens. See WithCookies.
	Cookies []*http.Cookie
}

var _ Screen = (*firstScreen)(nil)

func (s *firstScreen) String() string {
	return "first screen"
}

func (s *firstScreen) CurrentPageMatches(_ context.Context) bool {
	return true
}

func (s *firstScreen) Do(ctx context.Context) error {
	if s.URL == "" {
		return &MissingOptionError{Option: "WithURL"}
	}

	err := seedCookies(ctx, s.URL, s.Cookies)
	if err != nil {
		return fmt.Errorf("seed cookies: %w", err)
	}

	err = chromedp.Navigate(s.URL).Do(ctx)
	if err != nil {
		return fmt.Errorf("navigate: %w", err)
	}

	return nil
}

func (s *firstScreen) ShouldWaitForResponse() bool {
	return true
}
