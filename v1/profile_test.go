package digiposte_test

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/holyhope/digiposte-go-sdk/v1"
)

var _ = ginkgo.Describe("Profile", func() {
	ginkgo.Context("Authenticated", func() {
		ginkgo.It("should return a valid profile", func(ctx ginkgo.SpecContext) {
			profile, err := digiposteClient.GetProfile(ctx, digiposte.ProfileModeDefault)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(profile).ToNot(gomega.BeNil())
			gomega.Expect(profile.UserInfo.Email).ToNot(gomega.BeEmpty())
		})
	})
})
