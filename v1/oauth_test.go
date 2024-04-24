package digiposte_test

import (
	"net/http"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/onsi/gomega/gstruct"
	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login/noop"
	"github.com/holyhope/digiposte-go-sdk/v1"
)

var _ = ginkgo.FDescribe("Oauth", func() {
	var (
		server *ghttp.Server
		client *digiposte.Client
	)

	ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
		server = ghttp.NewServer()
		ginkgo.DeferCleanup(server.Close)

		c, err := digiposte.NewAuthenticatedClient(ctx, http.DefaultClient, &digiposte.Config{
			APIURL:      server.URL(),
			DocumentURL: server.URL(),
			Credentials: nil,
			LoginMethod: &noop.LoginMethod{
				Token: &oauth2.Token{
					AccessToken:  "token",
					TokenType:    "Bearer",
					RefreshToken: "refresh",
					Expiry:       time.Now().Add(time.Minute),
				},
				Cookies: nil,
			},
			SessionListener: nil,
			PreviousSession: nil,
		})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(c).ToNot(gomega.BeNil())

		client = c
	})

	ginkgo.Describe("NewAuthenticatedClient", func() {
		ginkgo.It("Should return a new client", func(ctx ginkgo.SpecContext) {
			expiresAt := time.Now().Add(time.Minute)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/rest/security/token"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &digiposte.AccessToken{
						Token:               "token",
						ExpiresAt:           expiresAt,
						IsTokenConsolidated: true,
					}),
				),
			)

			gomega.Expect(client.AccessToken(ctx)).To(gstruct.PointTo(gstruct.MatchAllFields(gstruct.Fields{
				"Token":               gomega.Equal("token"),
				"ExpiresAt":           gomega.BeTemporally("~", expiresAt, time.Millisecond),
				"IsTokenConsolidated": gomega.BeTrue(),
			})))

			gomega.Expect(server.ReceivedRequests()).Should(gomega.HaveLen(1))
		})
	})
})
