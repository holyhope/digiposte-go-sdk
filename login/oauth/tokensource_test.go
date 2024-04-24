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

var _ = ginkgo.Describe("TokenSource", func() {
	var tokenSource *oauth.TokenSource
	var expiry time.Time

	ginkgo.BeforeEach(func() {
		var nbTokens atomic.Int32
		var nbCalls atomic.Int32

		expiry = time.Now().Add(time.Hour)

		tokenSource = &oauth.TokenSource{
			LoginMethod: &MockedLoginMethod{
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

	ginkgo.It("Should generate a token", func() {
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

var _ = ginkgo.Describe("CombinedTokenSources", func() {
	var tokenSources oauth.CombinedTokenSources

	ginkgo.Describe("Withouth any token sources", func() {
		ginkgo.It("Should returns an error", func() {
			_, err := tokenSources.Token()
			gomega.Expect(err).To(gomega.MatchError(oauth.ErrNoTokenSources))
		})
	})

	ginkgo.Describe("With only errored token source", func() {
		ginkgo.BeforeEach(func() {
			tokenSources = oauth.CombinedTokenSources{
				&MockedTokenSource{
					TokenSource: func() (*oauth2.Token, error) {
						return nil, &idError{ID: 1}
					},
				},
				&MockedTokenSource{
					TokenSource: func() (*oauth2.Token, error) {
						return nil, &idError{ID: 2}
					},
				},
			}
		})

		ginkgo.It("Should returns all errors", func() {
			_, err := tokenSources.Token()
			gomega.Expect(err).To(gomega.MatchError("source 1: error 1\nsource 2: error 2"))
		})
	})

	ginkgo.Describe("With one valid token source", func() {
		var expiry time.Time
		ginkgo.BeforeEach(func() {
			expiry = time.Now().Add(time.Hour)
			tokenSources = oauth.CombinedTokenSources{
				&MockedTokenSource{
					TokenSource: func() (*oauth2.Token, error) {
						return nil, &idError{ID: 1}
					},
				},
				&MockedTokenSource{
					TokenSource: func() (*oauth2.Token, error) {
						return nil, &idError{ID: 2}
					},
				},
				&MockedTokenSource{
					TokenSource: func() (*oauth2.Token, error) {
						return &oauth2.Token{
							AccessToken:  "access-token",
							TokenType:    "token-type",
							RefreshToken: "refresh-token",
							Expiry:       expiry,
						}, nil
					},
				},
				&MockedTokenSource{
					TokenSource: func() (*oauth2.Token, error) {
						panic("should not be called")
					},
				},
			}
		})

		ginkgo.It("Should returns the token without errors", func() {
			gomega.Expect(tokenSources.Token()).To(gomega.Equal(&oauth2.Token{
				AccessToken:  "access-token",
				TokenType:    "token-type",
				RefreshToken: "refresh-token",
				Expiry:       expiry,
			}))
		})
	})
})

// MockedLoginMethod is a mock of login.Method.
type MockedLoginMethod struct {
	LoginMethod func(ctx context.Context, creds *login.Credentials) (*oauth2.Token, []*http.Cookie, error)
}

var _ login.Method = (*MockedLoginMethod)(nil)

func (mm *MockedLoginMethod) Login(
	ctx context.Context,
	creds *login.Credentials,
) (*oauth2.Token, []*http.Cookie, error) {
	return mm.LoginMethod(ctx, creds)
}

// MockedTokenSource is a mock of oauth2.TokenSource.
type MockedTokenSource struct {
	TokenSource func() (*oauth2.Token, error)
}

var _ login.Method = (*MockedLoginMethod)(nil)

func (mts *MockedTokenSource) Token() (*oauth2.Token, error) {
	return mts.TokenSource()
}

type idError struct {
	ID int
}

func (e *idError) Error() string {
	return fmt.Sprintf("error %d", e.ID)
}
