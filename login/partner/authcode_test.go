package partner_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login/partner"
)

const (
	testRedirectURL  = "https://caller.invalid/callback"
	testAuthURL      = "https://example.invalid/v3/authorize"
	testRefreshToken = "test-refresh-token"
)

func newAuthCodeConfig(tokenURL, okapiKey string) partner.AuthorizationCodeConfig {
	return partner.AuthorizationCodeConfig{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		Endpoint: partner.Endpoint{
			AuthURL:   testAuthURL,
			TokenURL:  tokenURL,
			AuthStyle: 0,
		},
		RedirectURL: testRedirectURL,
		OkapiKey:    okapiKey,
		Scopes:      nil,
		Timeout:     0,
	}
}

var _ = ginkgo.Describe("NewAuthorizationCodeFlow", func() {
	ginkgo.When("OkapiKey is missing", func() {
		ginkgo.It("returns an error", func() {
			_, err := partner.NewAuthorizationCodeFlow(newAuthCodeConfig("https://example.invalid/v3/token", ""))
			gomega.Expect(err).To(gomega.MatchError(partner.ErrMissingOkapiKey))
		})
	})

	ginkgo.When("building two independent flows", func() {
		ginkgo.It("generates distinct State and CodeVerifier values for each", func() {
			cfg := newAuthCodeConfig("https://example.invalid/v3/token", testOkapiKey)

			flowA, err := partner.NewAuthorizationCodeFlow(cfg)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			flowB, err := partner.NewAuthorizationCodeFlow(cfg)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(flowA.State).NotTo(gomega.BeEmpty())
			gomega.Expect(flowA.CodeVerifier).NotTo(gomega.BeEmpty())
			gomega.Expect(flowA.State).NotTo(gomega.Equal(flowB.State))
			gomega.Expect(flowA.CodeVerifier).NotTo(gomega.Equal(flowB.CodeVerifier))
		})
	})

	ginkgo.Describe("AuthCodeURL", func() {
		ginkgo.It("embeds client_id, redirect_uri, state and a PKCE S256 challenge", func() {
			cfg := newAuthCodeConfig("https://example.invalid/v3/token", testOkapiKey)

			flow, err := partner.NewAuthorizationCodeFlow(cfg)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			parsed, err := url.Parse(flow.AuthCodeURL())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			query := parsed.Query()
			gomega.Expect(query.Get("client_id")).To(gomega.Equal(cfg.ClientID))
			gomega.Expect(query.Get("redirect_uri")).To(gomega.Equal(cfg.RedirectURL))
			gomega.Expect(query.Get("state")).To(gomega.Equal(flow.State))
			gomega.Expect(query.Get("code_challenge")).To(gomega.Equal(oauth2.S256ChallengeFromVerifier(flow.CodeVerifier)))
			gomega.Expect(query.Get("code_challenge_method")).To(gomega.Equal("S256"))
		})
	})

	ginkgo.Describe("Exchange", func() {
		ginkgo.When("the supplied state does not match the flow's state", func() {
			ginkgo.It("fails without making any HTTP request", func() {
				var called atomic.Bool

				server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
					called.Store(true)

					resp.WriteHeader(http.StatusOK)
				}))
				defer server.Close()

				flow, err := partner.NewAuthorizationCodeFlow(newAuthCodeConfig(server.URL, testOkapiKey))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				_, err = flow.Exchange(context.Background(), "some-code", "wrong-state")
				gomega.Expect(err).To(gomega.MatchError(partner.ErrStateMismatch))
				gomega.Expect(called.Load()).To(gomega.BeFalse())
			})
		})

		ginkgo.When("the code exchange succeeds", func() {
			ginkgo.It("returns a token, sending the Okapi key and the original code_verifier", func() {
				var gotOkapiKey string

				var gotForm url.Values

				server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
					gotOkapiKey = req.Header.Get(partner.OkapiKeyHeader)

					gomega.Expect(req.ParseForm()).To(gomega.Succeed())
					gotForm = req.Form

					resp.Header().Set("Content-Type", "application/json")
					_, _ = resp.Write([]byte(`{"access_token":"` + testAccessToken +
						`","token_type":"Bearer","expires_in":3600,"refresh_token":"` + testRefreshToken + `"}`))
				}))
				defer server.Close()

				flow, err := partner.NewAuthorizationCodeFlow(newAuthCodeConfig(server.URL, testOkapiKey))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				token, err := flow.Exchange(context.Background(), "the-code", flow.State)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(token.AccessToken).To(gomega.Equal(testAccessToken))
				gomega.Expect(token.RefreshToken).To(gomega.Equal(testRefreshToken))

				gomega.Expect(gotOkapiKey).To(gomega.Equal(testOkapiKey))
				gomega.Expect(gotForm.Get("code_verifier")).To(gomega.Equal(flow.CodeVerifier))
				gomega.Expect(gotForm.Get("code")).To(gomega.Equal("the-code"))
			})
		})

		ginkgo.When("the Partner API rejects the code", func() {
			ginkgo.It("returns an error and no token", func() {
				server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
					resp.WriteHeader(http.StatusBadRequest)
					_, _ = resp.Write([]byte(`{"error":"invalid_grant","error_description":"invalid token"}`))
				}))
				defer server.Close()

				flow, err := partner.NewAuthorizationCodeFlow(newAuthCodeConfig(server.URL, testOkapiKey))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				token, err := flow.Exchange(context.Background(), "expired-code", flow.State)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(token).To(gomega.BeNil())
			})
		})
	})
})
