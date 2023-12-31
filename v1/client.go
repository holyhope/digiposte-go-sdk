package digiposte

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	DefaultAPIURL      = "https://api.digiposte.fr/api"
	DefaultDocumentURL = "https://secure.digiposte.fr"

	StagingAPIURL      = "https://api.interop.digiposte.io/api"
	StagingDocumentURL = "https://secure.interop.digiposte.io"
)

// Client is a Digiposte client.
type Client struct {
	apiURL      string
	documentURL string
	client      *http.Client
}

// NewClient creates a new Digiposte client.
func NewClient(client *http.Client) *Client {
	return NewCustomClient(DefaultAPIURL, DefaultDocumentURL, client)
}

// NewClient creates a new Digiposte client.
func NewCustomClient(apiURL, documentURL string, client *http.Client) *Client {
	return &Client{
		apiURL:      strings.TrimRight(apiURL, "/"),
		documentURL: strings.TrimRight(documentURL, "/"),
		client:      client,
	}
}

func (c *Client) apiRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, c.apiURL+path, body) //nolint:wrapcheck
}

const TrashDirName = "trash"

// ID represents an internal digiposte ID.
type ID string

// CloseBodyError is an error returned when the body of a response cannot be closed.
type CloseBodyError struct {
	Err           error
	OriginalError error
}

func (e *CloseBodyError) Error() string {
	return fmt.Sprintf("close body: %v", e.Err)
}

func (e *CloseBodyError) Unwrap() error {
	if e.OriginalError != nil {
		return e.OriginalError
	}

	return e.Err
}

// RequestError is an error returned when the API returns an error.
type RequestError struct {
	ErrorCode string                 `json:"error"`
	ErrorDesc string                 `json:"error_description,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("%s (%s)", e.ErrorDesc, e.ErrorCode)
}

func (c *Client) checkResponse(response *http.Response, expectedStatus int) error {
	if response.StatusCode != expectedStatus {
		var typedError RequestError

		content, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("HTTP %s: failed to read response body: %w", response.Status, err)
		}

		if err := json.Unmarshal(content, &typedError); err != nil {
			return &RequestError{
				ErrorCode: response.Status,
				ErrorDesc: string(content),
				Context: map[string]interface{}{
					"content-type": response.Header.Get("Content-Type"),
				},
			}
		}

		return fmt.Errorf("%s: %w", response.Status, &typedError)
	}

	return nil
}

// Trash move trashes the given documents and folders to the trash.
func (c *Client) Trash(ctx context.Context, documentIDs, folderIDs []ID) error {
	body, err := json.Marshal(map[string]interface{}{
		"document_ids": documentIDs,
		"folder_ids":   folderIDs,
	})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/file/tree/trash", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	queryParams := req.URL.Query()
	queryParams.Set("check", "false")
	req.URL.RawQuery = queryParams.Encode()

	req.Header.Set("Content-Type", "application/json")

	return c.call(req, nil)
}

// Delete deletes permanently the given documents and folders.
func (c *Client) Delete(ctx context.Context, documentIDs, folderIDs []ID) error {
	body, err := json.Marshal(map[string]interface{}{
		"document_ids": documentIDs,
		"folder_ids":   folderIDs,
	})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/file/tree/delete", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	return c.call(req, nil)
}

// Move moves the given documents and folders to the given destination.
func (c *Client) Move(ctx context.Context, destinationID ID, documentIDs, folderIDs []ID) error {
	body, err := json.Marshal(map[string]interface{}{
		"document_ids": documentIDs,
		"folder_ids":   folderIDs,
	})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	endpoint := "/v3/file/tree/move?to=" + url.QueryEscape(string(destinationID))

	req, err := c.apiRequest(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return c.call(req, nil)
}

func (c *Client) call(req *http.Request, result interface{}) (finalErr error) {
	response, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request %q: %w", req.URL, err)
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			finalErr = &CloseBodyError{Err: err, OriginalError: finalErr}
		}
	}()

	expectedStatus := http.StatusOK
	if result == nil {
		expectedStatus = http.StatusNoContent
	}

	if err := c.checkResponse(response, expectedStatus); err != nil {
		return fmt.Errorf("request to %q: %w", req.URL, err)
	}

	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// Logout logs out the user.
func (c *Client) Logout(ctx context.Context) error {
	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/profile/logout", nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	return c.call(req, nil)
}
