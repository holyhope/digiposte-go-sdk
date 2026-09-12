package partner_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login/partner"
)

const (
	testClientID     = "test-client-id"
	testClientSecret = "test-client-secret"
	testOkapiKey     = "test-okapi-key" //nolint:gosec
	testAccessToken  = "test-access-token"
)

func tokenResponseHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"access_token":"` + testAccessToken + `","token_type":"Bearer","expires_in":3600}`))
}

var _ = ginkgo.Describe("NewClientCredentialsSource", func() {
	ginkgo.When("OkapiKey is missing", func() {
		ginkgo.It("returns an error without making any HTTP request", func() {
			var called atomic.Bool

			server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
				called.Store(true)

				resp.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			_, err := partner.NewClientCredentialsSource(context.Background(), partner.ClientCredentialsConfig{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				Endpoint: partner.Endpoint{
					AuthURL:   "",
					TokenURL:  server.URL,
					AuthStyle: 0,
				},
				OkapiKey: "",
				Scopes:   nil,
				Timeout:  0,
			})

			gomega.Expect(err).To(gomega.MatchError(partner.ErrMissingOkapiKey))
			gomega.Expect(called.Load()).To(gomega.BeFalse())
		})
	})

	ginkgo.When("the grant succeeds", func() {
		var (
			gotOkapiKey  string
			gotAuthHdr   string
			requestCount atomic.Int32
			server       *httptest.Server
		)

		ginkgo.BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				requestCount.Add(1)

				gotOkapiKey = req.Header.Get(partner.OkapiKeyHeader)
				gotAuthHdr = req.Header.Get("Authorization")

				tokenResponseHandler(resp, req)
			}))
		})

		ginkgo.AfterEach(func() {
			server.Close()
		})

		ginkgo.It("returns a valid token, sending the Okapi key and HTTP Basic credentials", func() {
			source, err := partner.NewClientCredentialsSource(context.Background(), partner.ClientCredentialsConfig{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				Endpoint: partner.Endpoint{
					AuthURL:   "",
					TokenURL:  server.URL,
					AuthStyle: 0,
				},
				OkapiKey: testOkapiKey,
				Scopes:   nil,
				Timeout:  0,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			token, err := source.Token()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(token.AccessToken).To(gomega.Equal(testAccessToken))
			gomega.Expect(token.Valid()).To(gomega.BeTrue())

			gomega.Expect(gotOkapiKey).To(gomega.Equal(testOkapiKey))

			wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(testClientID+":"+testClientSecret))
			gomega.Expect(gotAuthHdr).To(gomega.Equal(wantAuth), "AuthStyle must default to HTTP Basic")
			gomega.Expect(requestCount.Load()).To(gomega.Equal(int32(1)))
		})
	})

	ginkgo.When("AuthStyle is explicitly set to send credentials in the request body", func() {
		ginkgo.It("puts client_id/client_secret in the form instead of the Authorization header", func() {
			var gotAuthHdr string

			var gotBodyClientID string

			server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
				gotAuthHdr = req.Header.Get("Authorization")

				_ = req.ParseForm()
				gotBodyClientID = req.PostForm.Get("client_id")

				tokenResponseHandler(resp, req)
			}))
			defer server.Close()

			source, err := partner.NewClientCredentialsSource(context.Background(), partner.ClientCredentialsConfig{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				Endpoint: partner.Endpoint{
					AuthURL:   "",
					TokenURL:  server.URL,
					AuthStyle: oauth2.AuthStyleInParams,
				},
				OkapiKey: testOkapiKey,
				Scopes:   nil,
				Timeout:  0,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = source.Token()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(gotAuthHdr).To(gomega.BeEmpty())
			gomega.Expect(gotBodyClientID).To(gomega.Equal(testClientID))
		})
	})

	ginkgo.When("the Partner API rejects the credentials", func() {
		ginkgo.It("fails without silently retrying in the request body", func() {
			var requestCount atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
				requestCount.Add(1)
				resp.WriteHeader(http.StatusUnauthorized)
				_, _ = resp.Write([]byte(`{"error":"invalid_client","error_description":"bad credentials"}`))
			}))
			defer server.Close()

			source, err := partner.NewClientCredentialsSource(context.Background(), partner.ClientCredentialsConfig{
				ClientID:     testClientID,
				ClientSecret: "wrong-secret",
				Endpoint: partner.Endpoint{
					AuthURL:   "",
					TokenURL:  server.URL,
					AuthStyle: 0,
				},
				OkapiKey: testOkapiKey,
				Scopes:   nil,
				Timeout:  0,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			token, err := source.Token()
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(token).To(gomega.BeNil())
			gomega.Expect(requestCount.Load()).To(gomega.Equal(int32(1)))
		})
	})

	ginkgo.When("the context is cancelled while a token request is in flight", func() {
		ginkgo.It("aborts the request instead of hanging", func() {
			release := make(chan struct{})

			server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
				<-release

				resp.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			defer close(release)

			ctx, cancel := context.WithCancel(context.Background())

			source, err := partner.NewClientCredentialsSource(ctx, partner.ClientCredentialsConfig{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				Endpoint: partner.Endpoint{
					AuthURL:   "",
					TokenURL:  server.URL,
					AuthStyle: 0,
				},
				OkapiKey: testOkapiKey,
				Scopes:   nil,
				Timeout:  time.Minute,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			done := make(chan error, 1)

			go func() {
				_, tokenErr := source.Token()
				done <- tokenErr
			}()

			cancel()

			gomega.Eventually(done, 5*time.Second).Should(
				gomega.Receive(gomega.MatchError(gomega.ContainSubstring("context canceled"))),
			)
		})
	})

	ginkgo.When("the token request exceeds its configured Timeout", func() {
		ginkgo.It("fails instead of waiting indefinitely", func() {
			release := make(chan struct{})

			server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
				<-release

				resp.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			defer close(release)

			source, err := partner.NewClientCredentialsSource(context.Background(), partner.ClientCredentialsConfig{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				Endpoint: partner.Endpoint{
					AuthURL:   "",
					TokenURL:  server.URL,
					AuthStyle: 0,
				},
				OkapiKey: testOkapiKey,
				Scopes:   nil,
				Timeout:  50 * time.Millisecond,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err = source.Token()
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})
