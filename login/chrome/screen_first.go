package chrome

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)

type firstScreen struct {
	URL string
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

	if err := chromedp.Navigate(s.URL).Do(ctx); err != nil {
		return fmt.Errorf("navigate: %w", err)
	}

	return nil
}

func (s *firstScreen) ShouldWaitForResponse() bool {
	return true
}

/*
func injectCookies(currentURL *url.URL, cookies []*http.Cookie) chromedp.Action {
	var injectCookies chromedp.Tasks

	for _, cookie := range cookies {
		injectCookies = append(injectCookies, chromeCookie(currentURL, cookie))
	}

	return injectCookies
}

func chromeCookie(u *url.URL, cookie *http.Cookie) *network.SetCookieParams {
	expire := cdp.TimeSinceEpoch(cookie.Expires)

	var sameSite network.CookieSameSite

	switch cookie.SameSite {
	case http.SameSiteLaxMode:
		sameSite = network.CookieSameSiteLax
	case http.SameSiteStrictMode:
		sameSite = network.CookieSameSiteStrict
	case http.SameSiteNoneMode, http.SameSiteDefaultMode:
		sameSite = network.CookieSameSiteNone
	}

	return network.SetCookie(cookie.Name, cookie.Value).
		WithDomain(u.Hostname()).
		WithPath(u.Path).
		WithURL(u.String()).
		WithExpires(&expire).
		WithHTTPOnly(cookie.HttpOnly).
		WithSecure(cookie.Secure).
		WithSameSite(sameSite)
}
*/
