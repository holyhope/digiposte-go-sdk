package chrome

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// ErrCookieDomainMismatch is returned when a cookie passed to WithCookies is
// explicitly scoped (via its Domain attribute) to a host that does not match
// the configured login URL.
var ErrCookieDomainMismatch = errors.New("cookie domain does not match login url")

// seedCookies applies cookies to the browser's cookie jar for rawURL's host,
// before any navigation happens, so that a still-authenticated site session
// (obtained previously and persisted by the caller, see WithCookies) is
// recognized by the site without driving the interactive login screens
// again. If cookies is empty, it does nothing.
//
// It is an error for any cookie to be explicitly scoped, via its Domain
// attribute, to a host that does not match rawURL's host: seeding a cookie
// that could never legitimately apply to the page about to be loaded most
// likely indicates a caller mistake (e.g. reusing cookies captured against a
// different URL or environment), so this fails loudly instead of silently
// dropping or misapplying it.
func seedCookies(ctx context.Context, rawURL string, cookies []*http.Cookie) error {
	if len(cookies) == 0 {
		return nil
	}

	target, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}

	host := target.Hostname()

	tasks := make(chromedp.Tasks, 0, len(cookies))

	for _, cookie := range cookies {
		if cookie.Domain != "" && !cookieDomainMatches(host, cookie.Domain) {
			return fmt.Errorf("%w: cookie %q is scoped to %q, login url host is %q",
				ErrCookieDomainMismatch, cookie.Name, cookie.Domain, host)
		}

		tasks = append(tasks, setCookieAction(target, cookie))
	}

	err = tasks.Do(ctx)
	if err != nil {
		return fmt.Errorf("set cookies: %w", err)
	}

	return nil
}

// cookieDomainMatches reports whether host is covered by a cookie's Domain
// attribute, following the usual cookie domain-matching rule: an exact
// match, or host being a subdomain of the (possibly dot-prefixed) cookie
// domain.
func cookieDomainMatches(host, cookieDomain string) bool {
	cookieDomain = strings.TrimPrefix(cookieDomain, ".")

	return host == cookieDomain || strings.HasSuffix(host, "."+cookieDomain)
}

func setCookieAction(target *url.URL, cookie *http.Cookie) *network.SetCookieParams {
	var sameSite network.CookieSameSite

	switch cookie.SameSite {
	case http.SameSiteLaxMode:
		sameSite = network.CookieSameSiteLax
	case http.SameSiteStrictMode:
		sameSite = network.CookieSameSiteStrict
	case http.SameSiteNoneMode, http.SameSiteDefaultMode:
		sameSite = network.CookieSameSiteNone
	}

	params := network.SetCookie(cookie.Name, cookie.Value).
		WithDomain(target.Hostname()).
		WithPath(cmp.Or(cookie.Path, "/")).
		WithURL(target.String()).
		WithHTTPOnly(cookie.HttpOnly).
		WithSecure(cookie.Secure).
		WithSameSite(sameSite)

	if !cookie.Expires.IsZero() {
		expires := cdp.TimeSinceEpoch(cookie.Expires)
		params = params.WithExpires(&expires)
	}

	return params
}
