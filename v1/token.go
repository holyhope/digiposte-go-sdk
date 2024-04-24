package digiposte

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/holyhope/digiposte-go-sdk/internal/utils"
)

type AccessToken struct {
	Token               string    `json:"access_token"`
	ExpiresAt           time.Time `json:"expires_at"`
	IsTokenConsolidated bool      `json:"is_token_consolidated"`
}

func (t *AccessToken) UnmarshalJSON(data []byte) error {
	var aux struct {
		Token               string  `json:"access_token"`
		ExpiresAt           float64 `json:"expires_at"`
		IsTokenConsolidated bool    `json:"is_token_consolidated"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	t.Token = aux.Token
	t.ExpiresAt = utils.UnixFloat2Time(aux.ExpiresAt)
	t.IsTokenConsolidated = aux.IsTokenConsolidated

	return nil
}

func (t *AccessToken) MarshalJSON() ([]byte, error) {
	aux := struct {
		Token               string  `json:"access_token"`
		ExpiresAt           float64 `json:"expires_at"`
		IsTokenConsolidated bool    `json:"is_token_consolidated"`
	}{
		Token:               t.Token,
		ExpiresAt:           utils.Time2UnixFloat(t.ExpiresAt),
		IsTokenConsolidated: t.IsTokenConsolidated,
	}

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	return data, nil
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

func (t *AppToken) UnmarshalJSON(data []byte) error {
	var aux struct {
		ExpiresAt float64 `json:"app_expires_at"`
		Token     string  `json:"app_access_token"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	t.ExpiresAt = utils.UnixFloat2Time(aux.ExpiresAt)
	t.Token = aux.Token

	return nil
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
