package oauth_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/internal/utils"
	"github.com/holyhope/digiposte-go-sdk/login"
	"github.com/holyhope/digiposte-go-sdk/login/chrome"
	"github.com/holyhope/digiposte-go-sdk/login/oauth"
)

func ExampleTokenSource() {
	path, err := utils.GetChrome(context.Background())
	if err != nil {
		panic(fmt.Errorf("get chrome: %w", err))
	}

	loginMethod, err := chrome.New(
		chrome.WithURL(os.Getenv("DIGIPOSTE_URL")),
		chrome.WithRefreshFrequency(500*time.Millisecond),
		chrome.WithScreenShortOnError(),
		chrome.WithTimeout(3*time.Minute),
		chrome.WithBinary(path),
	)
	if err != nil {
		panic(fmt.Errorf("new chrome login method: %w", err))
	}

	oauthTokenSource := oauth2.ReuseTokenSource(nil, &oauth.TokenSource{
		LoginMethod: loginMethod,
		Credentials: &login.Credentials{
			Username:  os.Getenv("DIGIPOSTE_USERNAME"),
			Password:  os.Getenv("DIGIPOSTE_PASSWORD"),
			OTPSecret: os.Getenv("DIGIPOSTE_OTP_SECRET"),
		},
		Listener: func(token *oauth2.Token, _ []*http.Cookie) {
			fmt.Printf("Token updated: %s\n", token.Type())
		},
	})

	token, err := oauthTokenSource.Token()
	if err != nil {
		panic(fmt.Errorf("get token: %w", err))
	}

	fmt.Printf("Token valid: %v\n", token.Valid())

	// Output:
	// Token updated: Bearer
	// Token valid: true
}
