package digiposte

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

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

	req.Header.Set("Content-Type", "application/json")

	var result SearchDocumentsResult

	return &result, c.call(req, &result)
}

// Document represents a document.
type Document struct {
	InternalID     ID        `json:"id"`
	Name           string    `json:"filename"`
	CreatedAt      time.Time `json:"creation_date"`
	Size           int64     `json:"size"`
	MimeType       string    `json:"mimetype"`
	FolderID       string    `json:"folder_id"`
	Location       string    `json:"location"`
	Shared         bool      `json:"shared"`
	Read           bool      `json:"read"`
	HealthDocument bool      `json:"health_document"`
	UserTags       []string  `json:"user_tags"`
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

// DocumentContent returns the content of a document.
func (c *Client) DocumentContent(ctx context.Context, internalID ID) (
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
			if finalErr == nil {
				finalErr = &CloseBodyError{Err: err}
			} else {
				finalErr = errors.Join(finalErr, &CloseBodyError{Err: err})
			}
		}
	}()

	contentType = response.Header.Get("Content-Type")

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
	LocationInbox Location = iota // INBOX
	LocationSafe                  // SAFE
	LocationTrash                 // TRASH
)

// SearchDocuments searches for documents in the given locations.
func (c *Client) SearchDocuments(ctx context.Context, internalID ID, locations ...Location) (
	*SearchDocumentsResult,
	error,
) {
	if len(locations) == 0 {
		locations = []Location{LocationInbox, LocationSafe}
	}

	var body interface{}
	if len(locations) == 1 && locations[0] == LocationTrash {
		body = map[string]interface{}{
			"folder_id": internalID,
			"trash":     true,
		}
	} else {
		locationsStr := make([]string, len(locations))
		for i, l := range locations {
			locationsStr[i] = l.String()
		}

		body = map[string]interface{}{
			"folder_id": internalID,
			"locations": locationsStr,
		}
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

	req.Header.Set("Content-Type", "application/json")

	var result SearchDocumentsResult

	return &result, c.call(req, &result)
}

// RenameDocument renames a document.
func (c *Client) RenameDocument(ctx context.Context, internalID ID, name string) (*Document, error) {
	endpoint := "/v3/document/" + url.PathEscape(string(internalID)) + "/rename/" + url.PathEscape(name)

	req, err := c.apiRequest(ctx, http.MethodPut, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	var document Document

	return &document, c.call(req, &document)
}

// CopyDocuments copies the given documents in the same folder.
func (c *Client) CopyDocuments(ctx context.Context, documentIDs []ID) (*SearchDocumentsResult, error) {
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

	req.Header.Set("Content-Type", "application/json")

	var result SearchDocumentsResult

	return &result, c.call(req, &result)
}

// MultiTag adds the given tags to the given documents.
func (c *Client) MultiTag(ctx context.Context, tags map[ID][]string) error {
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

	req.Header.Set("Content-Type", "application/json")

	return c.call(req, nil)
}

//go:generate stringer -type=DocumentType -trimprefix=DocumentType

type DocumentType int8

const (
	DocumentTypeBasic DocumentType = iota
	DocumentTypeHealth
)

// CreateDocument creates a document.
func (c *Client) CreateDocument(
	ctx context.Context,
	folderID ID,
	name string,
	data io.Reader,
	docType DocumentType,
) (document *Document, finalErr error) {
	var buf bytes.Buffer

	formWriter := multipart.NewWriter(&buf)
	defer func(formWriter *multipart.Writer) {
		if err := formWriter.Close(); err != nil {
			err := fmt.Errorf("close writer: %w", err)

			if finalErr == nil {
				finalErr = err
			} else {
				finalErr = errors.Join(finalErr, err)
			}
		}
	}(formWriter)

	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/document", &buf)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	// Copy the actual file content to the target destination
	if err := populateUploadForm(formWriter, docType, folderID, name, data); err != nil {
		return nil, fmt.Errorf("populate form: %w", err)
	}

	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	document = new(Document)

	return document, c.call(req, document)
}

func populateUploadForm(
	formWriter *multipart.Writer,
	docType DocumentType,
	folderID ID,
	name string,
	data io.Reader,
) (finalErr error) {
	if err := formWriter.WriteField("health_document", strconv.FormatBool(docType == DocumentTypeHealth)); err != nil {
		return fmt.Errorf("write health_document: %w", err)
	}

	if err := formWriter.WriteField("folder_id", string(folderID)); err != nil {
		return fmt.Errorf("write folder_id: %w", err)
	}

	if err := formWriter.WriteField("title", name); err != nil {
		return fmt.Errorf("write title: %w", err)
	}

	documentUploadStream, err := formWriter.CreateFormFile("archive", name)
	if err != nil {
		return fmt.Errorf("create archive file: %w", err)
	}

	sizeStream, err := formWriter.CreateFormField("archive_size")
	if err != nil {
		return fmt.Errorf("create archive_size field: %w", err)
	}

	go func(documentUploadStream, sizeStream io.Writer, content io.Reader) {
		if err := upload(documentUploadStream, content, sizeStream); err != nil {
			err := fmt.Errorf("upload: %w", err)

			if finalErr == nil {
				finalErr = err
			} else {
				finalErr = errors.Join(finalErr, err)
			}
		}
	}(documentUploadStream, sizeStream, data)

	return nil
}

func upload(documentUploadStream io.Writer, content io.Reader, sizeStream io.Writer) error {
	size, err := io.Copy(documentUploadStream, content)
	if err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	if _, err := sizeStream.Write([]byte(strconv.FormatInt(size, 10))); err != nil {
		return fmt.Errorf("write size: %w", err)
	}

	return nil
}
