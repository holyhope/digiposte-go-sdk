package digiposte_test

import (
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/holyhope/digiposte-go-sdk/v1"
)

var _ = ginkgo.Describe("Document", func() {
	ginkgo.Describe("CreateDocument", func() {
		var document *digiposte.Document

		ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
			gomega.Expect(digiposteClient.Trash(ctx, []digiposte.DocumentID{document.InternalID}, nil)).To(gomega.Succeed())
			gomega.Expect(digiposteClient.Delete(ctx, []digiposte.DocumentID{document.InternalID}, nil)).To(gomega.Succeed())
		})

		ginkgo.Context("when the document does not exist", func() {
			ginkgo.It("should create a document", func(ctx ginkgo.SpecContext) {
				var err error

				document, err = digiposteClient.CreateDocument(ctx,
					digiposte.RootFolderID,
					ginkgo.CurrentSpecReport().FullText(),
					strings.NewReader("the content"),
					digiposte.DocumentTypeBasic,
				)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(document.InternalID).ToNot(gomega.BeEmpty())
			})
		})
	})
})
