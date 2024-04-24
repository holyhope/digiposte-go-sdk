package oauth_test

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login"
	"github.com/holyhope/digiposte-go-sdk/login/oauth"
)

var _ = ginkgo.FDescribe("Token", func() {
	var tokenSource *oauth.TokenSource
	var expiry time.Time

	ginkgo.BeforeEach(func() {
		var nbTokens atomic.Int32
		var nbCalls atomic.Int32

		expiry = time.Now().Add(time.Hour)

		tokenSource = &oauth.TokenSource{
			LoginMethod: &MockedMethod{
				LoginMethod: func(_ context.Context, creds *login.Credentials) (*oauth2.Token, []*http.Cookie, error) {
					gomega.Expect(creds).To(gomega.Equal(&login.Credentials{
						Username:  "username",
						Password:  "password",
						OTPSecret: "otp-secret",
					}))

					return &oauth2.Token{
						AccessToken:  fmt.Sprintf("access-token-%d", nbTokens.Add(1)),
						TokenType:    "token-type",
						RefreshToken: "refresh-token",
						Expiry:       expiry,
					}, nil, nil
				},
			},
			Credentials: &login.Credentials{
				Username:  "username",
				Password:  "password",
				OTPSecret: "otp-secret",
			},
			Listener: func(token *oauth2.Token, cookies []*http.Cookie) {
				gomega.Expect(token).To(gomega.Equal(&oauth2.Token{
					AccessToken:  fmt.Sprintf("access-token-%d", nbCalls.Add(1)),
					TokenType:    "token-type",
					RefreshToken: "refresh-token",
					Expiry:       expiry,
				}))
				gomega.Expect(cookies).To(gomega.BeEmpty())
			},
		}

		ginkgo.DeferCleanup(func() {
			gomega.Expect(nbTokens.Load()).To(gomega.Equal(nbCalls.Load()))
		})
	})

	ginkgo.It("should generate a token", func() {
		gomega.Expect(tokenSource.Token()).To(gomega.Equal(&oauth2.Token{
			AccessToken:  "access-token-1",
			TokenType:    "token-type",
			RefreshToken: "refresh-token",
			Expiry:       expiry,
		}))
		gomega.Expect(tokenSource.Token()).To(gomega.Equal(&oauth2.Token{
			AccessToken:  "access-token-2",
			TokenType:    "token-type",
			RefreshToken: "refresh-token",
			Expiry:       expiry,
		}))
	})
})

// MockedMethod is a mock of login.Method.
type MockedMethod struct {
	LoginMethod func(ctx context.Context, creds *login.Credentials) (*oauth2.Token, []*http.Cookie, error)
}

var _ login.Method = (*MockedMethod)(nil)

func (mm *MockedMethod) Login(ctx context.Context, creds *login.Credentials) (*oauth2.Token, []*http.Cookie, error) {
	return mm.LoginMethod(ctx, creds)
}
