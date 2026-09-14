package partner_test

import (
	"context"
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/holyhope/digiposte-go-sdk/login/partner"
)

// Sandbox client credentials, publicly documented by Digiposte
// (https://developer.laposte.fr/catalog-apis/digiposte@3) - not secret.
// Only the Okapi key (DIGIPOSTE_OKAPI_TOKEN) is a real credential.
const (
	liveSandboxClientID     = "okapi"
	liveSandboxClientSecret = "6mstVSvc38wq"                              //nolint:gosec
	liveSandboxTokenURL     = "https://api.laposte.fr/digiposte/v3/token" //nolint:gosec
)

var _ = ginkgo.Describe("NewClientCredentialsSource", ginkgo.Label("live"), func() {
	ginkgo.It("obtains a valid access token from the real Partner API sandbox", func() {
		okapiKey := os.Getenv("DIGIPOSTE_OKAPI_TOKEN")
		if okapiKey == "" {
			ginkgo.Skip("missing DIGIPOSTE_OKAPI_TOKEN")
		}

		source, err := partner.NewClientCredentialsSource(context.Background(), partner.ClientCredentialsConfig{
			ClientID:     liveSandboxClientID,
			ClientSecret: liveSandboxClientSecret,
			Endpoint: partner.Endpoint{
				AuthURL:   "",
				TokenURL:  liveSandboxTokenURL,
				AuthStyle: 0,
			},
			OkapiKey: okapiKey,
			Scopes:   nil,
			Timeout:  0,
		})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		token, err := source.Token()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(token.Valid()).To(gomega.BeTrue())
		gomega.Expect(token.AccessToken).ToNot(gomega.BeEmpty())
	})
})
