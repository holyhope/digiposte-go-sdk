package digiposte_test

import (
	"context"
	"fmt"

	"github.com/holyhope/digiposte-go-sdk/v1"
)

func ExampleClient_ListFolders() {
	ctx := context.Background()

	// Create a new authenticated HTTP client using the following environment variables:
	// - DIGIPOSTE_API
	// - DIGIPOSTE_URL
	// - DIGIPOSTE_USERNAME
	// - DIGIPOSTE_PASSWORD
	// - DIGIPOSTE_OTP_SECRET
	client, err := DigiposteClient(ctx)
	if err != nil {
		panic(fmt.Errorf("new digiposte client: %w", err))
	}

	/* Create folders */

	folder, err := client.CreateFolder(ctx, digiposte.RootFolderID, "digiposte-go-sdk ExampleClient_ListFolders")
	if err != nil {
		panic(fmt.Errorf("create folder: %w", err))
	}

	fmt.Printf("Folder %q created\n", folder.Name)

	/* List folders */

	folders, err := client.ListFolders(ctx)
	if err != nil {
		panic(fmt.Errorf("list folders: %w", err))
	}

	for _, f := range folders.Folders {
		if f.InternalID == folder.InternalID {
			fmt.Print("Folder found\n")
		}
	}

	/* Trash the folder */

	if err := client.Trash(ctx, nil, []digiposte.FolderID{folder.InternalID}); err != nil {
		panic(fmt.Errorf("trash: %w", err))
	}

	fmt.Printf("Folder %q trashed\n", folder.Name)

	/* Delete the folder */

	if err := client.Delete(ctx, nil, []digiposte.FolderID{folder.InternalID}); err != nil {
		panic(fmt.Errorf("delete: %w", err))
	}

	fmt.Printf("Folder %q deleted\n", folder.Name)
	// Output:
	// Folder "digiposte-go-sdk ExampleClient_ListFolders" created
	// Folder found
	// Folder "digiposte-go-sdk ExampleClient_ListFolders" trashed
	// Folder "digiposte-go-sdk ExampleClient_ListFolders" deleted
}

func ExampleClient_GetTrashedFolders() {
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

	/* Create folders */

	folder, err := client.CreateFolder(ctx, digiposte.RootFolderID, "digiposte-go-sdk ExampleClient_GetTrashedFolders")
	if err != nil {
		panic(fmt.Errorf("create folder: %w", err))
	}

	fmt.Printf("Folder %q created\n", folder.Name)

	/* Trash the folder */

	if err := client.Trash(ctx, nil, []digiposte.FolderID{folder.InternalID}); err != nil {
		panic(fmt.Errorf("trash: %w", err))
	}

	fmt.Printf("Folder %q trashed\n", folder.Name)

	/* Get trashed folders */

	folders, err := client.GetTrashedFolders(ctx)
	if err != nil {
		panic(fmt.Errorf("get trashed folders: %w", err))
	}

	for _, f := range folders.Folders {
		if f.InternalID == folder.InternalID {
			fmt.Print("Trashed folder found\n")
		}
	}

	/* Delete the folder */

	if err := client.Delete(ctx, nil, []digiposte.FolderID{folder.InternalID}); err != nil {
		panic(fmt.Errorf("delete: %w", err))
	}

	fmt.Printf("Folder %q deleted\n", folder.Name)

	// Output:
	// Folder "digiposte-go-sdk ExampleClient_GetTrashedFolders" created
	// Folder "digiposte-go-sdk ExampleClient_GetTrashedFolders" trashed
	// Trashed folder found
	// Folder "digiposte-go-sdk ExampleClient_GetTrashedFolders" deleted
}

func ExampleClient_RenameFolder() {
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

	/* Create folders */

	folder, err := client.CreateFolder(ctx, digiposte.RootFolderID, "digiposte-go-sdk ExampleClient_R3n4m4F0ld3r")
	if err != nil {
		panic(fmt.Errorf("create folder: %w", err))
	}

	fmt.Printf("Folder %q created\n", folder.Name)

	folder, err = client.RenameFolder(ctx, folder.InternalID, "digiposte-go-sdk ExampleClient_RenameFolder")
	if err != nil {
		panic(fmt.Errorf("rename folder: %w", err))
	}

	fmt.Printf("Folder renamed to %q\n", folder.Name)

	/* Trash the folder */

	if err := client.Trash(ctx, nil, []digiposte.FolderID{folder.InternalID}); err != nil {
		panic(fmt.Errorf("trash: %w", err))
	}

	fmt.Printf("Folder %q trashed\n", folder.Name)

	/* Delete the folder */

	if err := client.Delete(ctx, nil, []digiposte.FolderID{folder.InternalID}); err != nil {
		panic(fmt.Errorf("delete: %w", err))
	}

	fmt.Printf("Folder %q deleted\n", folder.Name)

	// Output:
	// Folder "digiposte-go-sdk ExampleClient_R3n4m4F0ld3r" created
	// Folder renamed to "digiposte-go-sdk ExampleClient_RenameFolder"
	// Folder "digiposte-go-sdk ExampleClient_RenameFolder" trashed
	// Folder "digiposte-go-sdk ExampleClient_RenameFolder" deleted
}

func ExampleClient_CreateFolder() {
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

	/* Create folders */

	folder, err := client.CreateFolder(ctx, digiposte.RootFolderID, "digiposte-go-sdk ExampleClient_CreateFolder")
	if err != nil {
		panic(fmt.Errorf("create folder: %w", err))
	}

	fmt.Printf("Folder %q created\n", folder.Name)

	if _, err := client.CreateFolder(ctx, folder.InternalID, "sub-folder"); err != nil {
		panic(fmt.Errorf("create folder: %w", err))
	}

	fmt.Print("Sub folder created\n")

	/* Trash the top folder */

	if err := client.Trash(ctx, nil, []digiposte.FolderID{folder.InternalID}); err != nil {
		panic(fmt.Errorf("trash: %w", err))
	}

	fmt.Printf("Folder %q trashed\n", folder.Name)

	/* Delete the top folder */

	if err := client.Delete(ctx, nil, []digiposte.FolderID{folder.InternalID}); err != nil {
		panic(fmt.Errorf("delete: %w", err))
	}

	fmt.Printf("Folder %q deleted\n", folder.Name)

	// Output:
	// Folder "digiposte-go-sdk ExampleClient_CreateFolder" created
	// Sub folder created
	// Folder "digiposte-go-sdk ExampleClient_CreateFolder" trashed
	// Folder "digiposte-go-sdk ExampleClient_CreateFolder" deleted
}
