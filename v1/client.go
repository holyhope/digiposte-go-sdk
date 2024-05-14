package digiposte

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/oauth2"

	login "github.com/holyhope/digiposte-go-sdk/login"
	"github.com/holyhope/digiposte-go-sdk/login/chrome"
	"github.com/holyhope/digiposte-go-sdk/login/oauth"
	"github.com/holyhope/digiposte-go-sdk/settings"
)

// Client is a Digiposte client.
type Client struct {
	*clientHelper

	apiURL      string
	documentURL string
}

// NewClient creates a new Digiposte client.
func NewClient(client *http.Client) *Client {
	return NewCustomClient(settings.DefaultAPIURL, settings.DefaultDocumentURL, client)
}

// Session represents a Digiposte session.
type Session struct {
	Token   *oauth2.Token
	Cookies []*http.Cookie
}

// Config is the configuration of a Digiposte client.
type Config struct {
	APIURL      string
	DocumentURL string

	LoginMethod login.Method
	Credentials *login.Credentials

	SessionListener func(session *Session)

	PreviousSession *Session
}

// SetupDefault sets up the default values of the configuration.
func (c *Config) SetupDefault(ctx context.Context) error {
	if c.APIURL == "" {
		c.APIURL = settings.DefaultAPIURL
	}

	if c.DocumentURL == "" {
		c.DocumentURL = settings.DefaultDocumentURL
	}

	if c.LoginMethod == nil {
		method, err := chrome.New(chrome.WithChromeVersion(ctx, 0, nil))
		if err != nil {
			return fmt.Errorf("new chrome login method: %w", err)
		}

		c.LoginMethod = method
	}

	if c.SessionListener == nil {
		c.SessionListener = func(_ *Session) {}
	}

	if c.PreviousSession == nil {
		c.PreviousSession = new(Session)
	}

	return nil
}

// NewAuthenticatedClient creates a new Digiposte client with the given credentials.
func NewAuthenticatedClient(ctx context.Context, httpClient *http.Client, config *Config) (*Client, error) {
	if config == nil {
		config = new(Config)
	}

	if err := config.SetupDefault(ctx); err != nil {
		return nil, fmt.Errorf("setup default config: %w", err)
	}

	documentURL, err := url.Parse(config.DocumentURL)
	if err != nil {
		return nil, fmt.Errorf("parse document URL: %w", err)
	}

	if httpClient.Jar == nil {
		httpClient.Jar, err = cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("new cookie jar: %w", err)
		}
	}

	httpClient.Jar.SetCookies(documentURL, config.PreviousSession.Cookies)

	tokenSource := &TokenSource{
		clientHelper: &clientHelper{client: httpClient},
		DocumentURL:  config.DocumentURL,
		GetContext:   nil,
	}

	authenticatedClient := new(http.Client)

	*authenticatedClient = *httpClient

	authenticatedClient.Transport = &oauth2.Transport{
		Base: httpClient.Transport,
		Source: oauth2.ReuseTokenSource(config.PreviousSession.Token, oauth.CombinedTokenSources{
			tokenSource,
			&oauth.TokenSource{
				LoginMethod: config.LoginMethod,
				Credentials: config.Credentials,
				Listener: func(token *oauth2.Token, cookies []*http.Cookie) {
					httpClient.Jar.SetCookies(documentURL, cookies)

					config.SessionListener(&Session{
						Token:   token,
						Cookies: cookies,
					})
				},
			},
		}),
	}

	return NewCustomClient(config.APIURL, config.DocumentURL, authenticatedClient), nil
}

// NewClient creates a new Digiposte client.
func NewCustomClient(apiURL, documentURL string, client *http.Client) *Client {
	if client.CheckRedirect == nil {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) == 0 {
				panic("should not happen")
			}

			if req.URL.Host != via[len(via)-1].URL.Host {
				return http.ErrUseLastResponse
			}

			return nil
		}
	}

	return &Client{
		clientHelper: &clientHelper{client: client},
		apiURL:       strings.TrimRight(apiURL, "/"),
		documentURL:  strings.TrimRight(documentURL, "/"),
	}
}

const JSONContentType = "application/json"

func (c *Client) apiRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.apiURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Accept", JSONContentType)
	req.Header.Set("Content-Type", JSONContentType)

	return req, nil
}

const TrashDirName = "trash"

// ID represents an internal digiposte ID.
type digiposteID string

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
type RequestErrors []struct {
	ErrorCode string                 `json:"error"`
	ErrorDesc string                 `json:"error_description,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

func (e *RequestErrors) Error() string {
	strs := make([]string, 0, len(*e))

	for _, err := range *e {
		strs = append(strs, fmt.Sprintf("%s: %s", err.ErrorCode, err.ErrorDesc))
	}

	return strings.Join(strs, "\n")
}

// Trash move trashes the given documents and folders to the trash.
func (c *Client) Trash(ctx context.Context, documentIDs []DocumentID, folderIDs []FolderID) error {
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

	return c.call(req, nil)
}

// Delete deletes permanently the given documents and folders.
func (c *Client) Delete(ctx context.Context, documentIDs []DocumentID, folderIDs []FolderID) error {
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
func (c *Client) Move(ctx context.Context, destID FolderID, documentIDs []DocumentID, folderIDs []FolderID) error {
	body, err := json.Marshal(map[string]interface{}{
		"document_ids": documentIDs,
		"folder_ids":   folderIDs,
	})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	endpoint := "/v3/file/tree/move?to=" + url.QueryEscape(string(destID))

	req, err := c.apiRequest(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	return c.call(req, nil)
}

// Logout logs out the user.
func (c *Client) Logout(ctx context.Context) error {
	req, err := c.apiRequest(ctx, http.MethodPost, "/v3/profile/logout", nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	return c.call(req, nil)
}

type clientHelper struct {
	client *http.Client
}

func (c *clientHelper) call(req *http.Request, result interface{}, expectedStatuses ...int) (finalErr error) {
	response, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request %q: %w", req.URL, err)
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			finalErr = &CloseBodyError{Err: err, OriginalError: finalErr}
		}
	}()

	if len(expectedStatuses) == 0 {
		if result == nil {
			expectedStatuses = []int{http.StatusNoContent}
		} else {
			expectedStatuses = []int{http.StatusOK}
		}
	}

	if err := c.checkResponse(response, expectedStatuses...); err != nil {
		return fmt.Errorf("%s to %q: %w", req.Method, req.URL, err)
	}

	if result == nil {
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *clientHelper) checkResponse(response *http.Response, expectedStatuses ...int) error {
	for _, expectedStatus := range expectedStatuses {
		if response.StatusCode == expectedStatus {
			return nil
		}
	}

	errs := new(RequestErrors)

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("HTTP %s: failed to read response body: %w", response.Status, err)
	}

	if err := json.Unmarshal(content, errs); err != nil {
		context := map[string]interface{}{
			"content":      content,
			"decode_error": err,
		}

		if contentType := response.Header.Get("Content-Type"); contentType != "" {
			context["content-type"] = contentType
		}

		return &RequestErrors{{
			ErrorCode: response.Status,
			ErrorDesc: "failed to decode error response",
			Context:   context,
		}}
	}

	return fmt.Errorf("HTTP %s: %w", response.Status, errs)
}
