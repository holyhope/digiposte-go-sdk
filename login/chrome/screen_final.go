package chrome

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/internal/utils"
)

type finalScreen struct {
	Token   *oauth2.Token
	Cookies []*http.Cookie
}

var _ Screen = (*finalScreen)(nil)

func (s *finalScreen) String() string {
	return "final screen"
}

func (s *finalScreen) CurrentPageMatches(ctx context.Context) bool {
	var currentLocation string

	if err := chromedp.Run(ctx,
		chromedp.Location(&currentLocation),
	); err != nil {
		errorLogger(ctx).Printf("run: %v\n", err)

		return false
	}

	currentURL, err := url.Parse(currentLocation)
	if err != nil {
		errorLogger(ctx).Printf("parse current location: %v\n", err)

		return false
	}

	return currentURL.Path == "/home"
}

type InvalidTokenError struct {
	Token *oauth2.Token
}

func (e *InvalidTokenError) Error() string {
	return fmt.Sprintf("invalid token: %v", e.Token)
}

func (s *finalScreen) Do(ctx context.Context) error {
	token := new(oauth2.Token)

	var expiryStr string

	infoLogger(ctx).Println("Fetching token from browser...")

	if err := (&chromedp.Tasks{
		chromedp.Poll(`sessionStorage.getItem("access_token")`, &token.AccessToken),
		// access_expires_at returns the current time, is this a bug?
		chromedp.Poll(`sessionStorage.getItem("app_expires_at")`, &expiryStr),
	}).Do(ctx); err != nil {
		return fmt.Errorf("fetch token from browser: %w", err)
	}

	expiry, err := utils.UnixString2Time(expiryStr)
	if err != nil {
		return fmt.Errorf("parse app_expires_at: %w", err)
	}

	token.Expiry = expiry

	if !token.Valid() {
		return &InvalidTokenError{Token: token}
	}

	var (
		cookies    []*http.Cookie
		currentURL string
	)

	infoLogger(ctx).Println("Fetching cookies from browser...")

	if err := (&chromedp.Tasks{
		chromedp.Location(&currentURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromeCookies, err := network.GetCookies().Do(ctx)
			if err != nil {
				return fmt.Errorf("get cookies: %w", err)
			}

			cookies = make([]*http.Cookie, 0, len(chromeCookies))
			for _, v := range chromeCookies {
				cookies = append(cookies, convertCookie(v))
			}

			return nil
		}),
	}).Do(ctx); err != nil {
		return fmt.Errorf("fetch cookies from browser: %w", err)
	}

	infoLogger(ctx).Printf("%d cookies fetched from %q\n", len(cookies), currentURL)

	s.Token = token
	s.Cookies = cookies

	return nil
}

func (s *finalScreen) ShouldWaitForResponse() bool {
	return false
}

func convertCookie(cookie *network.Cookie) *http.Cookie {
	var sameSite http.SameSite

	switch cookie.SameSite {
	case network.CookieSameSiteLax:
		sameSite = http.SameSiteLaxMode
	case network.CookieSameSiteStrict:
		sameSite = http.SameSiteStrictMode
	case network.CookieSameSiteNone:
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteDefaultMode
	}

	return &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		Expires:  utils.UnixFloat2Time(cookie.Expires),
		Secure:   cookie.Secure,
		HttpOnly: cookie.HTTPOnly,
		SameSite: sameSite,
		Raw:      fmt.Sprintf("%s=%s", cookie.Name, cookie.Value),

		RawExpires: "",
		MaxAge:     0,
		Unparsed:   nil,
	}
}
