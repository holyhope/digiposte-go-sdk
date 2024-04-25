package digiposte_test

import (
	"context"
	"net/http"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/onsi/gomega/gstruct"

	"github.com/holyhope/digiposte-go-sdk/v1"
)

var _ = ginkgo.Describe("Oauth", func() {
	var (
		server    *ghttp.Server
		expiresAt time.Time
	)

	ginkgo.BeforeEach(func() {
		server = ghttp.NewServer()
		ginkgo.DeferCleanup(server.Close)

		expiresAt = time.Now().Add(time.Minute)
	})

	ginkgo.Describe("TokenSource", func() {
		var tokenSource *digiposte.TokenSource

		ginkgo.BeforeEach(func() {
			tokenSource = digiposte.NewTokenSource(http.DefaultClient, server.URL(), nil)
		})

		ginkgo.It("Should work", func(ctx ginkgo.SpecContext) {
			tokenSource.GetContext = func() context.Context {
				return ctx
			}

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

			gomega.Expect(tokenSource.Token()).To(gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"AccessToken":  gomega.Equal("token"),
				"TokenType":    gomega.Equal("Bearer"),
				"RefreshToken": gomega.Equal(""),
				"Expiry":       gomega.BeTemporally("~", expiresAt, time.Millisecond),
			})))

			gomega.Expect(server.ReceivedRequests()).Should(gomega.HaveLen(1))
		})
	})
})
