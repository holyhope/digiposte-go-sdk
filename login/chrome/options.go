package chrome

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	cu "github.com/Davincible/chromedp-undetected"

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

var errNegativeFreq = fmt.Errorf("frequency must be positive")

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

var errNegativeTimeout = fmt.Errorf("timeout must be positive")

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