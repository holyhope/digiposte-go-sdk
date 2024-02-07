package digiposte_test

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/holyhope/digiposte-go-sdk/v1"
)

var _ = ginkgo.Describe("Share", func() {
	ginkgo.Context("Authenticated", func() {
		ginkgo.Describe("Create share", func() {
			var share *digiposte.Share

			ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
				gomega.Expect(digiposteClient.DeleteShare(ctx, share.InternalID)).To(gomega.Succeed())
			})

			ginkgo.It("should return a valid share", func(ctx ginkgo.SpecContext) {
				t := time.Now()
				name := ginkgo.CurrentSpecReport().FullText()

				var err error

				share, err = digiposteClient.CreateShare(ctx, t, t.Add(2*time.Minute), name, "a password")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(share).ToNot(gomega.BeNil())
				gomega.Expect(share.InternalID).ToNot(gomega.BeEmpty())
			})
		})
	})
})
