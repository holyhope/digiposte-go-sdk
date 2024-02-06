package digiposte_test

import (
	"context"
	"embed"
	"fmt"
	"io"

	"github.com/holyhope/digiposte-go-sdk/v1"
)

//go:embed testdata/document.txt
var testData embed.FS

// ListFolders returns all folders at the root.
func Example() { //nolint:funlen
	ctx := context.Background()

	/* Create a new authenticated HTTP client using the following environment variables:
	 * - DIGIPOSTE_API
	 * - DIGIPOSTE_URL
	 * - DIGIPOSTE_USERNAME
	 * - DIGIPOSTE_PASSWORD
	 * - DIGIPOSTE_OTP_SECRET
	 */

	client, err := DigiposteClient(ctx)
	if err != nil {
		panic(fmt.Errorf("new digiposte client: %w", err))
	}

	/* Handle the cleanup of the created folder and document */

	var (
		folders   []digiposte.FolderID
		documents []digiposte.DocumentID
	)

	defer func(ctx context.Context) {
		if err := client.Trash(ctx, documents, folders); err != nil {
			panic(fmt.Errorf("trash: %w", err))
		}

		fmt.Printf("Trashed %d document(s)\n", len(documents))

		if err := client.Delete(ctx, documents, folders); err != nil {
			panic(fmt.Errorf("cleanup: %w", err))
		}

		fmt.Printf("Permanently deleted %d document(s)\n", len(documents))
	}(context.Background())

	/* Create a folder */

	folder, err := client.CreateFolder(ctx, digiposte.RootFolderID, "digiposte-go-sdk Example")
	if err != nil {
		panic(fmt.Errorf("create folder: %w", err))
	}

	folders = append(folders, folder.InternalID)

	fmt.Printf("Folder %q created\n", folder.Name)

	/* Create a document */

	document, err := testData.Open("testdata/document.txt")
	if err != nil {
		panic(fmt.Errorf("open testdata file: %w", err))
	}

	stat, err := document.Stat()
	if err != nil {
		panic(fmt.Errorf("stat testdata file: %w", err))
	}

	doc, err := client.CreateDocument(ctx, folder.InternalID, stat.Name(), document, digiposte.DocumentTypeBasic)
	if err != nil {
		panic(fmt.Errorf("create document: %w", err))
	}

	documents = append(documents, doc.InternalID)

	fmt.Printf("Document %q created\n", doc.Name)

	/* Get document content */

	contentReader, contentType, err := client.DocumentContent(ctx, doc.InternalID)
	if err != nil {
		panic(fmt.Errorf("get document content: %w", err))
	}

	fmt.Printf("Document content type: %s\n", contentType)

	content, err := io.ReadAll(contentReader)
	if err != nil {
		panic(fmt.Errorf("read document content: %w", err))
	}

	fmt.Printf("Document size: %d bytes\n", len(content))

	// Output:
	// Folder "digiposte-go-sdk Example" created
	// Document "document.txt" created
	// Document content type: text/plain;charset=UTF-8
	// Document size: 134 bytes
	// Trashed 1 document(s)
	// Permanently deleted 1 document(s)
}
