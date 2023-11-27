package digiposte

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
func NewClient(client *http.Client) (*Client, error) {
	return NewCustomClient(DefaultAPIURL, DefaultDocumentURL, client)
}

// NewClient creates a new Digiposte client.
func NewCustomClient(apiURL, documentURL string, client *http.Client) (*Client, error) {
	return &Client{
		apiURL:      strings.TrimRight(apiURL, "/"),
		documentURL: strings.TrimRight(documentURL, "/"),
		client:      client,
	}, nil
}

func (c *Client) apiRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, c.apiURL+path, body) //nolint:wrapcheck
}

const TrashDirName = "trash"

type Profile struct {
	Status                string `json:"status"`
	SpaceUsed             int64  `json:"space_used"`
	SpaceFree             int64  `json:"space_free"`
	SpaceMax              int64  `json:"space_max"`
	SpaceNotComputed      int64  `json:"space_not_computed"`
	AuthorName            string `json:"author_name"`
	ShareSpaceStatus      string `json:"share_space_status"`
	HasOfflineModeAbility bool   `json:"has_offline_mode_ability"`

	Offer struct {
		SubscriptionDate time.Time `json:"subscription_date"`
	} `json:"offer"`

	UserInfo struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Login     string `json:"login"`
		Email     string `json:"primary_email"`
	} `json:",inline"`
}

// ID represents an internal digiposte ID.
type ID string

//go:generate stringer -type=ProfileMode -linecomment

type ProfileMode int

const (
	ProfileModeDefault            ProfileMode = iota // default
	ProfileModeNoSpaceConsumption                    // without_space_consumption
)

// GetProfile returns the profile of the user.
func (c *Client) GetProfile(ctx context.Context, mode fmt.Stringer) (profile *Profile, finalErr error) {
	req, err := c.apiRequest(ctx, http.MethodGet, "/v4/profile", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	if mode != ProfileModeDefault {
		queryParams := req.URL.Query()
		queryParams.Set("mode", mode.String())
		req.URL.RawQuery = queryParams.Encode()
	}

	profile = new(Profile)

	return profile, c.call(req, profile)
}

// ProfileSafeSize represents the usage of the safe.
type ProfileSafeSize struct {
	ActualSafeSize int64 `json:"actual_safe_size"`
}

// GetProfileSafeSize returns the usage of the safe.
func (c *Client) GetProfileSafeSize(ctx context.Context) (currentSize *ProfileSafeSize, finalErr error) {
	req, err := c.apiRequest(ctx, http.MethodGet, "/v4/profile/safe/size", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	currentSize = new(ProfileSafeSize)

	return currentSize, c.call(req, currentSize)
}

// CloseBodyError is an error returned when the body of a response cannot be closed.
type CloseBodyError struct {
	Err error
}

func (e *CloseBodyError) Error() string {
	return fmt.Sprintf("close body: %v", e.Err)
}

func (e *CloseBodyError) Unwrap() error {
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
func (c *Client) Trash(ctx context.Context, documentIDs, folderIDs []ID) (finalErr error) {
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
func (c *Client) Delete(ctx context.Context, documentIDs, folderIDs []ID) (finalErr error) {
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
func (c *Client) Move(ctx context.Context, destinationID ID, documentIDs, folderIDs []ID) (finalErr error) {
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
			if finalErr == nil {
				finalErr = &CloseBodyError{Err: err}
			} else {
				finalErr = errors.Join(finalErr, &CloseBodyError{Err: err})
			}
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
