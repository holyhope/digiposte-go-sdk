package digiposte_test

import (
	"net/http"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login/noop"
	"github.com/holyhope/digiposte-go-sdk/settings"
	"github.com/holyhope/digiposte-go-sdk/v1"
)

var _ = ginkgo.Describe("Client", func() {
	ginkgo.Describe("NewClient", func() {
		ginkgo.It("Should return a new client", func() {
			gomega.Expect(digiposte.NewClient(nil)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("NewAuthenticatedClient", func() {
		ginkgo.It("Should return a new client", func(ctx ginkgo.SpecContext) {
			gomega.Expect(digiposte.NewAuthenticatedClient(ctx, http.DefaultClient, &digiposte.Config{
				APIURL:      "",
				DocumentURL: "",
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
			})).ToNot(gomega.BeNil())
		})
	})
})

var _ = ginkgo.Describe("Config", func() {
	ginkgo.Describe("SetupDefault", func() {
		ginkgo.It("Should work", func(ctx ginkgo.SpecContext) {
			var config digiposte.Config

			gomega.Expect(config.SetupDefault(ctx)).To(gomega.Succeed())
			gomega.Expect(config.APIURL).To(gomega.Equal(settings.DefaultAPIURL))
			gomega.Expect(config.DocumentURL).To(gomega.Equal(settings.DefaultDocumentURL))
			gomega.Expect(config.LoginMethod).ToNot(gomega.BeNil())
		})
	})
})
