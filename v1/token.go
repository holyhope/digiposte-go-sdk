package digiposte

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type AccessToken struct {
	Token               string    `json:"access_token"`
	ExpiresAt           time.Time `json:"expires_at"`
	IsTokenConsolidated bool      `json:"is_token_consolidated"`
}

// Token returns a actoken.
func (c *Client) AccessToken(ctx context.Context) (*AccessToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.documentURL+"/rest/security/token", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	result := new(AccessToken)

	return result, c.call(req, result)
}

type AppToken struct {
	Token     string    `json:"app_access_token"`
	ExpiresAt time.Time `json:"app_expires_at"`
}

// AppToken returns a application token.
func (c *Client) AppToken(ctx context.Context) (*AppToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.documentURL+"/rest/security/app-token", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	result := new(AppToken)

	return result, c.call(req, result)
}
