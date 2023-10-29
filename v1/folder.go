package digiposte

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Folder represents a Digiposte folder.
type Folder struct {
	InternalID    ID        `json:"id"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DocumentCount int64     `json:"document_count"`
	Folders       []*Folder `json:"folders"`
}

// SearchFoldersResult represents a search result for folders.
type SearchFoldersResult struct {
	Count      int64     `json:"count"`
	Index      int64     `json:"index"`
	MaxResults int64     `json:"max_results"`
	Folders    []*Folder `json:"folders"`
}

// ListFolders returns all folders at the root.
func (c *Client) ListFolders(ctx context.Context) (result *SearchFoldersResult, finalErr error) {
	req, err := c.apiRequest(ctx, http.MethodGet, "/v3/folders", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	result = new(SearchFoldersResult)

	return result, c.call(req, result)
}

// GetTrashedFolders returns all folders in the trash.
func (c *Client) GetTrashedFolders(ctx context.Context) (result *SearchFoldersResult, finalErr error) {
	req, err := c.apiRequest(ctx, http.MethodGet, "/v3/folders/"+TrashDirName, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	result = new(SearchFoldersResult)

	return result, c.call(req, result)
}

// RenameFolder renames a folder.
func (c *Client) RenameFolder(ctx context.Context, internalID ID, name string) (folder *Folder, finalErr error) {
	endpoint := "/v3/folder/" + url.PathEscape(string(internalID)) + "/rename/" + url.PathEscape(name)

	req, err := c.apiRequest(ctx, http.MethodPut, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	folder = new(Folder)

	return folder, c.call(req, folder)
}
