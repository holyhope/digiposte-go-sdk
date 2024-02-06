package digiconfig

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/holyhope/digiposte-go-sdk/settings"
)

const (
	APIURLKey      = "api_url"      // Configuration key for API URL
	DocumentURLKey = "document_url" // Configuration key for document URL
	UsernameKey    = "username"     // Configuration key for username
	PasswordKey    = "password"     // Configuration key for password
	OTPSecretKey   = "otp"          // Configuration key for otp
	CookiesKey     = "cookies"      // Configuration key for cookie
)

var (
	MustReveal  = func(s string) string { return s } //nolint:gochecknoglobals
	MustObscure = func(s string) string { return s } //nolint:gochecknoglobals
)

func DocumentURL(m Getter) string {
	val, ok := m.Get(DocumentURLKey)
	if !ok {
		return settings.DefaultDocumentURL
	}

	return val
}

func SetDocumentURL(m Setter, documentURL string) {
	m.Set(DocumentURLKey, documentURL)
}

func APIURL(m Getter) string {
	val, ok := m.Get(APIURLKey)
	if !ok {
		return settings.DefaultAPIURL
	}

	return val
}

func SetAPIURL(m Setter, apiURL string) {
	m.Set(APIURLKey, apiURL)
}

func Username(m Getter) string {
	val, _ := m.Get(UsernameKey)

	return val
}

func SetUsername(m Setter, username string) {
	m.Set(UsernameKey, username)
}

func Password(m Getter) string {
	val, _ := m.Get(PasswordKey)

	return MustReveal(val)
}

func SetPassword(m Setter, password string) {
	m.Set(PasswordKey, MustObscure(password))
}

func OTPSecret(m Getter) string {
	val, _ := m.Get(OTPSecretKey)

	return MustReveal(val)
}

func SetOTPSecret(m Setter, otpSecret string) {
	m.Set(OTPSecretKey, MustObscure(otpSecret))
}

func Cookies(m Getter) []*http.Cookie {
	val, ok := m.Get(CookiesKey)
	if !ok {
		return nil
	}

	var cypheredCookies []*http.Cookie
	if err := json.Unmarshal([]byte(val), &cypheredCookies); err != nil {
		panic(fmt.Errorf("unmarshal cookies: %w", err))
	}

	cookies := make([]*http.Cookie, 0, len(cypheredCookies))

	for _, cookie := range cypheredCookies {
		if err := cookie.Valid(); err != nil {
			panic(fmt.Errorf("invalid cookie %q: %w", cookie.Name, err))
		}

		cookie := *cookie
		cookie.Value = MustReveal(cookie.Value)
		cookies = append(cookies, &cookie)
	}

	return cookies
}

func SetCookies(setter Setter, cookies []*http.Cookie) error {
	cypheredCookies := make([]*http.Cookie, 0, len(cookies))

	for _, cookie := range cookies {
		cookie := *cookie
		cookie.Value = MustObscure(cookie.Value)
		cypheredCookies = append(cypheredCookies, &cookie)
	}

	cookiesBytes, err := json.Marshal(cypheredCookies)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	setter.Set(CookiesKey, string(cookiesBytes))

	return nil
}
