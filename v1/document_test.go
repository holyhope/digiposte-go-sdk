package digiposte_test

import (
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"

	"github.com/holyhope/digiposte-go-sdk/v1"
)

var _ = ginkgo.Describe("Document", func() {
	var document *digiposte.Document

	ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
		if document == nil {
			return
		}

		if err := digiposteClient.Trash(ctx, []digiposte.DocumentID{document.InternalID}, nil); err != nil {
			fmt.Fprintf(ginkgo.GinkgoWriter, "trash: %v\n", err)
		}

		if err := digiposteClient.Delete(ctx, []digiposte.DocumentID{document.InternalID}, nil); err != nil {
			fmt.Fprintf(ginkgo.GinkgoWriter, "delete: %v\n", err)
		}
	})

	ginkgo.Describe("CreateDocument", func() {
		ginkgo.Context("When the document does not exist", func() {
			ginkgo.It("Should create a document", func(ctx ginkgo.SpecContext) {
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

	ginkgo.Describe("TagDocument", func() {
		ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
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

		ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
			if err := digiposteClient.Trash(ctx, []digiposte.DocumentID{document.InternalID}, nil); err != nil {
				fmt.Fprintf(ginkgo.GinkgoWriter, "trash: %v\n", err)
			}

			if err := digiposteClient.Delete(ctx, []digiposte.DocumentID{document.InternalID}, nil); err != nil {
				fmt.Fprintf(ginkgo.GinkgoWriter, "delete: %v\n", err)
			}
		})

		ginkgo.It("Should tag the document", func(ctx ginkgo.SpecContext) {
			tag := digiposte.DocumentTag(strings.ReplaceAll(strings.ToLower(ginkgo.CurrentSpecReport().FullText()), " ", "-"))

			tagsBefore, err := digiposteClient.UserTags(ctx)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			countBefore, ok := tagsBefore.Tags[tag]
			if !ok {
				countBefore = 0
			}

			gomega.Expect(digiposteClient.MultiTag(ctx, map[digiposte.DocumentID][]digiposte.DocumentTag{document.InternalID: {
				tag,
			}})).To(gomega.Succeed())

			gomega.Eventually(func() *digiposte.UserTags {
				result, err := digiposteClient.UserTags(ctx)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				return result
			}).Should(gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Tags": gomega.HaveKeyWithValue(tag, countBefore+1),
			})))

			result, err := digiposteClient.SearchDocuments(ctx, digiposte.RootFolderID, digiposte.DocumentTaggedWith(tag))
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(result.Documents).To(gomega.ConsistOf(gstruct.PointTo(
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"InternalID": gomega.BeEquivalentTo(document.InternalID),
					"UserTags":   gomega.ContainElement(gomega.BeEquivalentTo(tag)),
				})),
			))
		})
	})

	ginkgo.Describe("DocumentContent", func() {
		ginkgo.Context("When the document does not exist", func() {
			ginkgo.BeforeEach(func() {
				document = new(digiposte.Document)
				document.InternalID = "i-do-not-exist"
			})

			ginkgo.It("Should return an error", func(ctx ginkgo.SpecContext) {
				_, _, err := digiposteClient.DocumentContent(ctx, document.InternalID)
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("When the document exist", func() {
			ginkgo.BeforeEach(func(ctx ginkgo.SpecContext) {
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

			ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
				gomega.Expect(digiposteClient.Trash(ctx, []digiposte.DocumentID{document.InternalID}, nil)).To(gomega.Succeed())
				gomega.Expect(digiposteClient.Delete(ctx, []digiposte.DocumentID{document.InternalID}, nil)).To(gomega.Succeed())
			})

			ginkgo.It("Should return the content and the right Content-Type", func(ctx ginkgo.SpecContext) {
				content, contentType, err := digiposteClient.DocumentContent(ctx, document.InternalID)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Eventually(gbytes.BufferReader(content)).Should(gbytes.Say("^the content$"))

				mediaType, params, err := mime.ParseMediaType(contentType)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(mediaType).To(gomega.Equal("text/plain"))
				gomega.Expect(params).To(gomega.HaveKeyWithValue(
					"charset", gomega.WithTransform(strings.ToLower, gomega.Equal("utf-8")),
				))
			})

			ginkgo.Context("With a non-authenticated client", func() {
				var client *digiposte.Client

				ginkgo.BeforeEach(func() {
					client = digiposte.NewClient(http.DefaultClient)
				})

				ginkgo.It("Should return an error", func(ctx ginkgo.SpecContext) {
					_, _, err := client.DocumentContent(ctx, document.InternalID)
					gomega.Expect(err).To(gomega.HaveOccurred())
					gomega.Expect(err).To(gstruct.PointTo(gomega.ContainElement(
						gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
							"ErrorCode": gomega.Equal("Unauthorized"),
							"ErrorDesc": gomega.Equal("Redirected to the login page."),
						}),
					)))
				})
			})
		})
	})
})
