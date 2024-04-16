package chrome

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/go-rod/rod/lib/launcher"

	login "github.com/holyhope/digiposte-go-sdk/login"
)

type Validatable interface {
	Validate() error
}

var chromeOpts = []cu.Option{ //nolint:gochecknoglobals
	func(c *cu.Config) {
		c.Language = "fr-FR"
	},
}

var errNegativeFreq = errors.New("frequency must be positive")

// WithRefreshFrequency sets the frequency at which the login process will be refreshed.
func WithRefreshFrequency(frequency time.Duration) login.Option { //nolint:ireturn
	return &withRefreshFrequency{Frequency: frequency}
}

type withRefreshFrequency struct {
	Frequency time.Duration
}

func (o *withRefreshFrequency) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.refreshFrequency = o.Frequency

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

func (o *withRefreshFrequency) Validate() error {
	if o.Frequency <= 0 {
		return &login.InvalidOptionError{
			Name: "WithRefreshFrequency",
			Err:  errNegativeFreq,
		}
	}

	return nil
}

var errNegativeTimeout = errors.New("timeout must be positive")

// WithTimeout sets the timeout for the login process.
func WithTimeout(timeout time.Duration) login.Option { //nolint:ireturn
	return &withTimeout{Timeout: timeout}
}

type withTimeout struct {
	Timeout time.Duration
}

func (o *withTimeout) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.timeout = o.Timeout

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

func (o *withTimeout) Validate() error {
	if o.Timeout <= 0 {
		return &login.InvalidOptionError{
			Name: "WithTimeout",
			Err:  errNegativeTimeout,
		}
	}

	return nil
}

var errEmptyURL = errors.New("url is empty")

// WithURL sets the URL to which the login process will be directed.
func WithURL(url string) login.Option { //nolint:ireturn
	return &withURL{URL: url}
}

type withURL struct {
	URL string
}

func (o *withURL) Validate() error {
	if o.URL == "" {
		return &login.InvalidOptionError{
			Name: "WithURL",
			Err:  errEmptyURL,
		}
	}

	return nil
}

func (o *withURL) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.url = o.URL

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

// WithCookies sets the cookies to be used for the login process.
func WithCookies(cookies []*http.Cookie) login.Option { //nolint:ireturn
	return &withCookies{Cookies: cookies}
}

type withCookies struct {
	Cookies []*http.Cookie
}

func (o *withCookies) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.cookies = o.Cookies

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

// WithScreenShortOnError sets the option to take a screenshot when an error occurs.
func WithScreenShortOnError() login.Option { //nolint:ireturn
	return &withScreenShortOnError{}
}

type withScreenShortOnError struct{}

func (o *withScreenShortOnError) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.screenShortOnError = true

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

type InvalidTypeOptionError struct {
	instance interface{}
}

func (e *InvalidTypeOptionError) Error() string {
	return fmt.Sprintf("invalid instance type: %T", e.instance)
}

const (
	ConfigURL       = "url"
	ConfigCookieJar = "cookie"
)

type MissingOptionError struct {
	Option string
}

func (e *MissingOptionError) Error() string {
	return fmt.Sprintf("missing option %q", e.Option)
}

func (e *MissingOptionError) Is(target error) bool {
	if target, ok := target.(*MissingOptionError); ok {
		return reflect.ValueOf(e.Option).Pointer() == reflect.ValueOf(target.Option).Pointer()
	}

	return false
}

type WithScreenshotError struct {
	Err        error
	Screenshot []byte
}

func (e *WithScreenshotError) Error() string {
	return e.Err.Error()
}

func (e *WithScreenshotError) Unwrap() error {
	return e.Err
}

func (e *WithScreenshotError) String() string {
	return fmt.Sprintf("error: %v\nscreenshot: %s", e.Err, byteCountSI(len(e.Screenshot)))
}

// WithLoggers sets the loggers to be used for the login process.
func WithLoggers(infoLgr, errorLgr *log.Logger) login.Option { //nolint:ireturn
	return &withLoggers{
		Info:  infoLgr,
		Error: errorLgr,
	}
}

type withLoggers struct {
	Info  *log.Logger
	Error *log.Logger
}

func (o *withLoggers) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		if o.Info != nil {
			chrome.infoLogger = o.Info
		}

		if o.Error != nil {
			chrome.errorLogger = o.Error
		}

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

// WithChromeVersion sets the version of chrome to be used for the login process.
func WithChromeVersion(ctx context.Context, revision int, client *http.Client) login.Option { //nolint:ireturn
	browser := launcher.NewBrowser()
	if ctx != nil {
		browser.Context = ctx
	}

	if revision > 0 {
		browser.Revision = revision
	}

	if client != nil {
		browser.HTTPClient = client
	}

	return &withChromeVersion{Browser: browser}
}

type withChromeVersion struct {
	Browser *launcher.Browser
}

func (o *withChromeVersion) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		o.Browser.Logger = chrome.infoLogger

		path, err := o.Browser.Get()
		if err != nil {
			return fmt.Errorf("get browser: %w", err)
		}

		if err := o.Browser.Validate(); err != nil {
			return fmt.Errorf("validate browser: %w", err)
		}

		chrome.binaryPath = path

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

// WithBinary sets the path to the chrome binary to be used for the login process.
func WithBinary(path string) login.Option { //nolint:ireturn
	return &withBinary{Path: path}
}

type withBinary struct {
	Path string
}

func (o *withBinary) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		if o.Path != "" {
			chrome.binaryPath = o.Path
		}

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}
