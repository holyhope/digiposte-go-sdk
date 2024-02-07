package oauth_test

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/onsi/gomega/ghttp"
	"golang.org/x/oauth2"

	login "github.com/holyhope/digiposte-go-sdk/login"
	digiconfig "github.com/holyhope/digiposte-go-sdk/login/config"
	configfakes "github.com/holyhope/digiposte-go-sdk/login/config/configfakes"
	oauth "github.com/holyhope/digiposte-go-sdk/login/oauth"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
)

const (
	ClientID     = "client-id"
	ClientSecret = "client-secret"
	Username     = "username"
	Password     = "password"
	OTPSecret    = "otp-secret"
)

var _ = Describe("Server", func() {
	var (
		setter      *configfakes.FakeSetter
		oauthServer *oauth.Server
		testServer  *ghttp.Server
		cfg         *oauth2.Config
	)

	BeforeEach(func() {
		testServer = ghttp.NewServer()
		DeferCleanup(testServer.Close)

		setter = &configfakes.FakeSetter{
			SetStub: func(key, value string) {
				Expect(key).To(Equal(digiconfig.CookiesKey))
				Expect(value).ToNot(BeEmpty())
			},
		}

		loginMethod := func(ctx context.Context, creds *login.Credentials) (*oauth2.Token, []*http.Cookie, error) {
			Expect(creds).To(Equal(&login.Credentials{
				Username:  Username,
				Password:  Password,
				OTPSecret: OTPSecret,
			}))

			return &oauth2.Token{
					AccessToken:  "access-token",
					TokenType:    "token-type",
					RefreshToken: "refresh-token",
					Expiry:       time.Now().Add(time.Hour),
				}, []*http.Cookie{{
					Name:   "cookie-name",
					Value:  "cookie-value",
					Path:   "/digi-test",
					Domain: "digi-test.fr",
				}}, nil
		}

		localServer, err := oauth.NewServer(
			setter,
			&oauth.Config{
				Addr:        ":0", // Random port
				Server:      server.NewConfig(),
				Logger:      log.New(GinkgoWriter, "", log.Lmsgprefix),
				LoginMethod: login.MethodFunc(loginMethod),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(localServer).ToNot(BeNil())

		Expect(localServer.RegisterUser(
			ClientID, ClientSecret, testServer.URL(),
			Username, Password, OTPSecret,
		)).To(Succeed())

		go func(server *oauth.Server) {
			defer GinkgoRecover()

			Expect(server.Start()).To(Succeed())
		}(localServer)

		DeferCleanup(func() {
			Expect(oauthServer.Shutdown(context.Background())).To(Succeed())
		})

		oauthServer = localServer

		cfg = &oauth2.Config{
			ClientID:     ClientID,
			ClientSecret: ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:       oauthServer.AuthorizeURL(),
				TokenURL:      oauthServer.TokenURL(),
				AuthStyle:     oauth2.AuthStyleInParams,
				DeviceAuthURL: "",
			},
			RedirectURL: testServer.URL(),
			Scopes:      nil,
		}
	})

	Context("When using a password", func() {
		It("Should fail with a bad password", func() {
			_, err := cfg.PasswordCredentialsToken(context.Background(), "username", "password")
			Expect(err).To(HaveOccurred())

			var targetErr *oauth2.RetrieveError

			if errors.As(err, &targetErr) {
				Expect(targetErr.Response.StatusCode).To(Equal(http.StatusForbidden))
			}
		})
	})

	It("Should be able to generate a token", func() {
		var code string

		req, err := http.NewRequest(http.MethodGet, cfg.AuthCodeURL("tests"), nil)
		Expect(err).ToNot(HaveOccurred())

		initialRequest := req.Clone(context.Background())

		testServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest(http.MethodGet, "/"),
			func(writer http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Referer")).To(Equal(initialRequest.URL.String()))

				query := req.URL.Query()
				Expect(query.Get("state")).To(Equal("tests"))

				code = query.Get("code")

				writer.WriteHeader(http.StatusNoContent)
			},
		))

		resp, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

		Expect(code).ToNot(BeEmpty())

		Expect(setter.Invocations()).To(BeEmpty())

		token, err := cfg.Exchange(context.Background(), code)
		Expect(err).ToNot(HaveOccurred())
		Expect(token.Valid()).To(BeTrue())

		Expect(setter.Invocations()).To(HaveKeyWithValue("Set", ConsistOf(
			ConsistOf(Equal(digiconfig.CookiesKey), Not(BeEmpty())),
		)))
	})
})
