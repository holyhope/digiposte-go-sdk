package digiposte

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type FolderID digiposteID

const RootFolderID FolderID = ""

// Folder represents a Digiposte folder.
type Folder struct {
	InternalID    FolderID  `json:"id"`
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
func (c *Client) ListFolders(ctx context.Context) (*SearchFoldersResult, error) {
	req, err := c.apiRequest(ctx, http.MethodGet, "/v3/folders", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	result := new(SearchFoldersResult)

	return result, c.call(req, result)
}

// GetTrashedFolders returns all folders in the trash.
func (c *Client) GetTrashedFolders(ctx context.Context) (*SearchFoldersResult, error) {
	req, err := c.apiRequest(ctx, http.MethodGet, "/v3/folders/"+TrashDirName, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	result := new(SearchFoldersResult)

	return result, c.call(req, result)
}

// RenameFolder renames a folder.
func (c *Client) RenameFolder(ctx context.Context, internalID FolderID, name string) (*Folder, error) {
	endpoint := "/v3/folder/" + url.PathEscape(string(internalID)) + "/rename/" + url.PathEscape(name)

	req, err := c.apiRequest(ctx, http.MethodPut, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	folder := new(Folder)

	return folder, c.call(req, folder)
}

// CreateFolder creates a folder.
func (c *Client) CreateFolder(ctx context.Context, parentID FolderID, name string) (*Folder, error) {
	body, err := json.Marshal(map[string]interface{}{
		"parent_id": parentID,
		"name":      name,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/folder", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	folder := new(Folder)

	return folder, c.call(req, folder)
}
