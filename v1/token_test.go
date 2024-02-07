package digiposte_test

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Access token", func() {
	ginkgo.Context("Authenticated", func() {
		ginkgo.It("should return a valid token", func(ctx ginkgo.SpecContext) {
			token, err := digiposteClient.AccessToken(ctx)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(token).ToNot(gomega.BeNil())
			gomega.Expect(token.Token).ToNot(gomega.BeEmpty())
		})
	})
})

var _ = ginkgo.Describe("App token", func() {
	ginkgo.Context("Authenticated", func() {
		ginkgo.It("should return a valid token", func(ctx ginkgo.SpecContext) {
			token, err := digiposteClient.AppToken(ctx)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(token).ToNot(gomega.BeNil())
			gomega.Expect(token.Token).ToNot(gomega.BeEmpty())
		})
	})
})
