package digiposte

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type DocumentID digiposteID

// GetTrashedDocuments returns all documents in the trash.
func (c *Client) GetTrashedDocuments(ctx context.Context) (*SearchDocumentsResult, error) {
	body, err := json.Marshal(map[string]interface{}{
		"user_removal": true,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/documents/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	queryParams := req.URL.Query()
	queryParams.Set("max_results", "1000")
	queryParams.Set("sort", "TITLE")
	req.URL.RawQuery = queryParams.Encode()

	var result SearchDocumentsResult

	return &result, c.call(req, &result)
}

// Document represents a document.
type Document struct {
	InternalID     DocumentID `json:"id"`
	Name           string     `json:"filename"`
	CreatedAt      time.Time  `json:"creation_date"`
	Size           int64      `json:"size"`
	MimeType       string     `json:"mimetype"`
	FolderID       string     `json:"folder_id"`
	Location       string     `json:"location"`
	Shared         bool       `json:"shared"`
	Read           bool       `json:"read"`
	HealthDocument bool       `json:"health_document"`
	UserTags       []string   `json:"user_tags"`
}

// ListDocuments returns all documents at the root.
func (c *Client) ListDocuments(ctx context.Context) (*SearchDocumentsResult, error) {
	req, err := c.apiRequest(ctx, http.MethodGet, "/v3/documents", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	var result SearchDocumentsResult

	return &result, c.call(req, &result)
}

type RedirectionError struct {
	Location string
}

func (e *RedirectionError) Error() string {
	return fmt.Sprintf("redirection stopped: %q", e.Location)
}

// DocumentContent returns the content of a document.
func (c *Client) DocumentContent(ctx context.Context, internalID DocumentID) ( //nolint:nonamedreturns
	contentBuffer io.ReadCloser,
	contentType string,
	finalErr error,
) {
	endpoint := c.documentURL + "/rest/content/document/" + url.PathEscape(string(internalID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("new request: %w", err)
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to request %q: %w", req.URL, err)
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			finalErr = &CloseBodyError{Err: err, OriginalError: finalErr}
		}
	}()

	contentType = response.Header.Get("Content-Type")

	if response.StatusCode == http.StatusFound {
		location, err := url.Parse(response.Header.Get("Location"))
		if err != nil {
			return nil, "", fmt.Errorf("parse location: %w", err)
		}

		if strings.HasSuffix(location.Path, "/v3/authorize") {
			return nil, "", &RequestErrors{{
				ErrorCode: http.StatusText(http.StatusUnauthorized),
				ErrorDesc: "Redirected to the login page.",
				Context:   map[string]interface{}{"response": response},
			}}
		}

		return nil, "", &RedirectionError{Location: location.String()}
	}

	if err := c.checkResponse(response, http.StatusOK); err != nil {
		return nil, contentType, fmt.Errorf("request to %q: %w", req.URL, err)
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, contentType, fmt.Errorf("failed to read response body: %w", err)
	}

	return io.NopCloser(bytes.NewReader(content)), contentType, nil
}

// SearchDocumentsResult represents the result of a search for documents.
type SearchDocumentsResult struct {
	Count      int64       `json:"count"`
	Index      int64       `json:"index"`
	MaxResults int64       `json:"max_results"`
	Documents  []*Document `json:"documents"`
}

//go:generate stringer -type=Location -linecomment

// Location represents a location of a document.
type Location int8

const (
	LocationInbox      Location = iota // INBOX
	LocationSafe                       // SAFE
	LocationTrashInbox                 // TRASH_INBOX
	LocationTrashSafe                  // TRASH_SAFE
)

// DocumentSearchOption represents an option for searching documents.
type DocumentSearchOption func(map[string]interface{})

// HealthDocuments returns only health documents.
func HealthDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["health"] = true
	}
}

// NotHealthDocuments returns only non-health documents.
func NotHealthDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["health"] = false
	}
}

// SharedDocuments returns only shared documents.
func SharedDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["document_shared"] = true
	}
}

// NotSharedDocuments returns only non-shared documents.
func NotSharedDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["document_shared"] = false
	}
}

// ReadDocuments returns only read documents.
func ReadDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["document_read"] = true
	}
}

// UnreadDocuments returns only unread documents.
func UnreadDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["document_read"] = false
	}
}

// CertifiedDocuments returns only certified documents.
func CertifiedDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["document_certified"] = true
	}
}

// NotCertifiedDocuments returns only non-certified documents.
func NotCertifiedDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["document_certified"] = false
	}
}

// FavoriteDocuments returns only favorite documents.
func FavoriteDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["favorite"] = true
	}
}

// NotFavoriteDocuments returns only non-favorite documents.
func NotFavoriteDocuments() DocumentSearchOption {
	return func(body map[string]interface{}) {
		body["favorite"] = false
	}
}

// OnlyDocumentLocatedAt returns only documents located at the given locations.
func OnlyDocumentLocatedAt(locations ...Location) DocumentSearchOption {
	locationsStr := make([]string, len(locations))
	for i, l := range locations {
		locationsStr[i] = l.String()
	}

	return func(body map[string]interface{}) {
		body["locations"] = locationsStr
	}
}

// SearchDocuments searches for documents in the given locations.
func (c *Client) SearchDocuments(ctx context.Context, internalID FolderID, options ...DocumentSearchOption) (
	*SearchDocumentsResult,
	error,
) {
	body := map[string]interface{}{
		"folder_id": internalID,
		"locations": []string{LocationInbox.String(), LocationSafe.String()},
	}

	for _, option := range options {
		option(body)
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/documents/search", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	queryParams := req.URL.Query()
	queryParams.Set("max_results", "1000")
	queryParams.Set("sort", "TITLE")
	req.URL.RawQuery = queryParams.Encode()

	var result SearchDocumentsResult

	return &result, c.call(req, &result)
}

// RenameDocument renames a document.
func (c *Client) RenameDocument(ctx context.Context, internalID DocumentID, name string) (*Document, error) {
	endpoint := "/v3/document/" + url.PathEscape(string(internalID)) + "/rename/" + url.PathEscape(name)

	req, err := c.apiRequest(ctx, http.MethodPut, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	var document Document

	return &document, c.call(req, &document)
}

// CopyDocuments copies the given documents in the same folder.
func (c *Client) CopyDocuments(ctx context.Context, documentIDs []DocumentID) (*SearchDocumentsResult, error) {
	body, err := json.Marshal(map[string]interface{}{
		"documents": documentIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/documents/copy", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	var result SearchDocumentsResult

	return &result, c.call(req, &result)
}

// MultiTag adds the given tags to the given documents.
func (c *Client) MultiTag(ctx context.Context, tags map[DocumentID][]string) error {
	body, err := json.Marshal(map[string]interface{}{
		"tags": tags,
	})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/documents/multi-tag", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	return c.call(req, nil)
}

//go:generate stringer -type=DocumentType -trimprefix=DocumentType

type DocumentType int8

const (
	DocumentTypeBasic DocumentType = iota
	DocumentTypeHealth
)

// CreateDocument creates a document.
func (c *Client) CreateDocument( //nolint:nonamedreturns
	ctx context.Context,
	folderID FolderID,
	name string,
	data io.Reader,
	docType DocumentType,
) (document *Document, finalErr error) {
	var buf bytes.Buffer

	formWriter, err := uploadForm(&buf, docType, folderID, name, data)
	if err != nil {
		return nil, fmt.Errorf("upload form: %w", err)
	}

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/document", &buf)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", formWriter.FormDataContentType())
	req.Header.Set("X-API-VERSION-MINOR", "2")
	req.Header.Set("Origin", "https://github.com/holyhope/digiposte-go-sdk")

	document = new(Document)

	return document, c.call(req, document)
}

func uploadForm(
	writer io.Writer,
	docType DocumentType,
	folderID FolderID,
	name string,
	data io.Reader,
) (*multipart.Writer, error) {
	formWriter := multipart.NewWriter(writer)

	if err := formWriter.WriteField("health_document", strconv.FormatBool(docType == DocumentTypeHealth)); err != nil {
		return formWriter, fmt.Errorf("write health_document: %w", err)
	}

	if folderID != "" {
		if err := formWriter.WriteField("folder_id", string(folderID)); err != nil {
			return formWriter, fmt.Errorf("write folder_id: %w", err)
		}
	}

	if err := formWriter.WriteField("title", name); err != nil {
		return formWriter, fmt.Errorf("write title: %w", err)
	}

	documentUploadStream, err := formWriter.CreateFormFile("archive", name)
	if err != nil {
		return formWriter, fmt.Errorf("create archive file: %w", err)
	}

	size, err := io.Copy(documentUploadStream, data)
	if err != nil {
		return formWriter, fmt.Errorf("copy archive file: %w", err)
	}

	if err := formWriter.WriteField("archive_size", strconv.FormatInt(size, 10)); err != nil {
		return formWriter, fmt.Errorf("write archive_size: %w", err)
	}

	if err := formWriter.Close(); err != nil {
		return formWriter, fmt.Errorf("close form writer: %w", err)
	}

	return formWriter, nil
}

// CloseWriterError represents an error during the closing of a writer.
type CloseWriterError struct {
	Err           error
	OriginalError error
}

func (e *CloseWriterError) Error() string {
	return fmt.Sprintf("upload: %v", e.Err)
}

func (e *CloseWriterError) Unwrap() error {
	if e.OriginalError != nil {
		return e.OriginalError
	}

	return e.Err
}

// UploadError represents an error during an upload.
type UploadError struct {
	Err           error
	OriginalError error
}

func (e *UploadError) Error() string {
	return fmt.Sprintf("upload: %v", e.Err)
}

func (e *UploadError) Unwrap() error {
	if e.OriginalError != nil {
		return e.OriginalError
	}

	return e.Err
}
