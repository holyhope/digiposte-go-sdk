package digiposte_test

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/holyhope/digiposte-go-sdk/v1"
)

var _ = ginkgo.Describe("Folder", func() {
	ginkgo.Describe("CreateFolder", func() {
		var folder *digiposte.Folder

		ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
			gomega.Expect(digiposteClient.Trash(ctx, nil, []digiposte.FolderID{folder.InternalID})).To(gomega.Succeed())
			gomega.Expect(digiposteClient.Delete(ctx, nil, []digiposte.FolderID{folder.InternalID})).To(gomega.Succeed())
		})

		ginkgo.Context("when the folder does not exist", func() {
			ginkgo.It("should create a folder", func(ctx ginkgo.SpecContext) {
				var err error

				folder, err = digiposteClient.CreateFolder(ctx, digiposte.RootFolderID, ginkgo.CurrentSpecReport().FullText())
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(folder.InternalID).ToNot(gomega.BeEmpty())
			})
		})
	})
})
