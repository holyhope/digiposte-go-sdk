package digiposte

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

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

//go:generate stringer -type=ProfileMode -linecomment

type ProfileMode int

const (
	ProfileModeDefault            ProfileMode = iota // default
	ProfileModeNoSpaceConsumption                    // without_space_consumption
)

// GetProfile returns the profile of the user.
func (c *Client) GetProfile(ctx context.Context, mode fmt.Stringer) (*Profile, error) {
	req, err := c.apiRequest(ctx, http.MethodGet, "/v4/profile", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	if mode != ProfileModeDefault {
		queryParams := req.URL.Query()
		queryParams.Set("mode", mode.String())
		req.URL.RawQuery = queryParams.Encode()
	}

	profile := new(Profile)

	return profile, c.call(req, profile)
}

// ProfileSafeSize represents the usage of the safe.
type ProfileSafeSize struct {
	ActualSafeSize int64 `json:"actual_safe_size"`
}

// GetProfileSafeSize returns the usage of the safe.
func (c *Client) GetProfileSafeSize(ctx context.Context) (*ProfileSafeSize, error) {
	req, err := c.apiRequest(ctx, http.MethodGet, "/v4/profile/safe/size", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	currentSize := new(ProfileSafeSize)

	return currentSize, c.call(req, currentSize)
}
