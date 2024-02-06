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

type ShareID digiposteID

const sharePrefix = "/v3/share/"

// Share represents a share.
type Share struct {
	InternalID     ShareID   `json:"id"`
	ShortID        string    `json:"short_id"`
	SecurityCode   string    `json:"security_code"`
	ShortURL       string    `json:"short_url"`
	Title          string    `json:"title"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	RecipientMails []string  `json:"recipient_mails"`
}

// Share creates a share for a specific time period, with a title and a security code.
func (c *Client) CreateShare(ctx context.Context, startDate, endDate time.Time, title, code string) (*Share, error) {
	body := map[string]interface{}{
		"start_date": startDate,
		"title":      title,
	}

	if !endDate.IsZero() {
		body["end_date"] = endDate
	}

	if code != "" {
		body["security_code"] = code
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/share", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	share := new(Share)

	return share, c.call(req, share)
}

// SetShareDocuments adds a document to a share.
func (c *Client) SetShareDocuments(ctx context.Context, shareID ShareID, documentIDs []DocumentID) error {
	body, err := json.Marshal(map[string]interface{}{
		"ids": documentIDs,
	})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	endpoint := sharePrefix + url.PathEscape(string(shareID)) + "/documents"

	req, err := c.apiRequest(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	return c.call(req, nil)
}

// ShareResult represents a share.
type ShareResult struct {
	SenderShares []Share `json:"senderShares"`
	ShareDatas   []Share `json:"shareDatas"`
}

// ListShares returns all shares.
func (c *Client) ListShares(ctx context.Context) (*ShareResult, error) {
	endpoint := "/v4/partner/user/shares"

	req, err := c.apiRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	share := new(ShareResult)

	return share, c.call(req, share)
}

// ShareResultWithDocuments represents a share with documents.
type ShareResultWithDocuments struct {
	ShareDataAndDocuments []struct {
		ShareData Share      `json:"share_data"`
		Documents []Document `json:"documents"`
	} `json:"share_data_and_documents"`
}

// ListSharesWithDocuments returns all shares with documents.
func (c *Client) ListSharesWithDocuments(ctx context.Context) (*ShareResultWithDocuments, error) {
	endpoint := "/v3/shares/with_documents"

	req, err := c.apiRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	result := new(ShareResultWithDocuments)

	return result, c.call(req, result)
}

// GetShareDocuments returns all documents of a share.
func (c *Client) GetShareDocuments(ctx context.Context, shareID ShareID) (*SearchDocumentsResult, error) {
	endpoint := sharePrefix + url.PathEscape(string(shareID)) + "/documents"

	req, err := c.apiRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	result := new(SearchDocumentsResult)

	return result, c.call(req, result)
}

// GetShare returns a share.
func (c *Client) GetShare(ctx context.Context, shareID ShareID) (*Share, error) {
	endpoint := sharePrefix + url.PathEscape(string(shareID))

	req, err := c.apiRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	share := new(Share)

	return share, c.call(req, share)
}

// DeleteShare deletes a share.
func (c *Client) DeleteShare(ctx context.Context, shareID ShareID) error {
	endpoint := sharePrefix + url.PathEscape(string(shareID))

	req, err := c.apiRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	return c.call(req, nil)
}
