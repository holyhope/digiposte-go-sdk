package chrome

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/chromedp"
	"golang.org/x/oauth2"

	login "github.com/holyhope/digiposte-go-sdk/login"
	"github.com/holyhope/digiposte-go-sdk/settings"
)

// New creates a new chrome login method.
// It requires chromedp to be installed.
func New(opts ...login.Option) (login.Method, error) { //nolint:ireturn
	for i, opt := range opts {
		if opt, ok := opt.(Validatable); ok {
			if err := opt.Validate(); err != nil {
				return nil, fmt.Errorf("validate option %d: %w", i, err)
			}
		}
	}

	return &chromeMethod{
		opts: opts,
	}, nil
}

type chromeMethod struct {
	opts []login.Option
}

var _ login.Method = (*chromeMethod)(nil)

// Login logs in to digiposte using chrome.
func (c *chromeMethod) Login(ctx context.Context, creds *login.Credentials) (*oauth2.Token, []*http.Cookie, error) {
	independentChromeCtx, chrome, cancel, err := c.newChromeLogin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("new chrome login: %w", err)
	}

	defer cancel()

	if err := chromedp.Run(independentChromeCtx); err != nil {
		return nil, nil, fmt.Errorf("init chrome: %w", err)
	}

	defer closeChrome(independentChromeCtx)

	pid := chromedp.FromContext(independentChromeCtx).Browser.Process().Pid

	infoLogger(independentChromeCtx).Printf("Chrome started. PID: %d...\n", pid)

	return chrome.login(ctx, independentChromeCtx, creds)
}

func (c *chromeMethod) String() string {
	return "chrome"
}

const (
	// DefaultRefreshFrequency is the default refresh frequency for the login process.
	DefaultRefreshFrequency = 1500 * time.Millisecond
)

func (c *chromeMethod) newChromeLogin(
	_ context.Context,
) (context.Context, *chromeLogin, context.CancelFunc, error) {
	chrome := &chromeLogin{
		refreshFrequency:   DefaultRefreshFrequency,
		url:                settings.DefaultDocumentURL,
		cookies:            nil,
		screenShortOnError: false,
		infoLogger:         log.Default(),
		errorLogger:        log.Default(),
		timeout:            0,
		binaryPath:         "",
	}

	for i, opt := range c.opts {
		if err := opt.Apply(chrome); err != nil {
			return nil, nil, nil, fmt.Errorf("apply option %d: %w", i, err)
		}
	}

	// Note: Do not inherit the context, so that we can cancel it independently.
	independentChromeCtx, cancelCtx := context.WithCancel(context.Background())

	independentChromeCtx = withInfoLogger(independentChromeCtx, chrome.infoLogger)
	independentChromeCtx = withErrorLogger(independentChromeCtx, chrome.errorLogger)

	independentChromeCtx, cancelChrome, err := cu.New(cu.NewConfig(append(chromeOpts,
		cu.WithContext(independentChromeCtx),
		cu.WithChromeBinary(chrome.binaryPath),
		func(c *cu.Config) {
			c.ContextOptions = append(c.ContextOptions,
				chromedp.WithErrorf(chrome.errorLogger.Printf),
				chromedp.WithLogf(chrome.infoLogger.Printf),
				chromedp.WithDebugf(func(_ string, _ ...interface{}) {
					// do nothing
				}),
			)
		},
	)...))
	if err != nil {
		cancelCtx()

		return nil, nil, nil, fmt.Errorf("new chromedp context: %w", err)
	}

	return independentChromeCtx, chrome, func() {
		cancelChrome()
		cancelCtx()
	}, nil
}

const cancellationTimeout = 5 * time.Second

func closeChrome(ctx context.Context) {
	proc := chromedp.FromContext(ctx).Browser.Process()

	ctx, cancel := context.WithTimeout(ctx, cancellationTimeout)
	defer cancel()

	if err := chromedp.Cancel(ctx); err != nil {
		lgr := errorLogger(ctx)

		lgr.Printf("Failed to cancel chrome: %v\n", err)

		if err := proc.Kill(); err != nil {
			lgr.Printf("Failed to kill chrome: %v\n", err)
		}
	}
}
