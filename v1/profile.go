package digiposte

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Serializable struct {
	Serializer   interface{} `json:"serializer"`
	Deserializer interface{} `json:"deserializer"`
}

type UserInfo struct {
	InternalID string      `json:"id"`
	Title      string      `json:"title"`
	FirstName  string      `json:"first_name"`
	LastName   string      `json:"last_name"`
	IDXiti     interface{} `json:"id_xiti"`
	Login      string      `json:"login"`
	Type       string      `json:"user_type"`
	Locale     string      `json:"locale"`
	Email      string      `json:"primaryEmail"`
}

type Offer struct {
	Serializable `json:",inline"`

	PID                         string    `json:"pid"`
	Type                        string    `json:"type"`
	MaxSafeSize                 int64     `json:"max_safe_size"`
	MaxCollectorsCount          int       `json:"max_nb_collectors"`
	ActualSafeSize              int64     `json:"actual_safe_size"`
	ActualCollectorsCount       int       `json:"actual_nb_collectors"`
	SubscriptionDate            time.Time `json:"subscription_date"`
	Price                       int64     `json:"price"`
	Frequency                   string    `json:"frequency"`
	CommercialName              string    `json:"commercial_name"`
	CanAddProCollectors         bool      `json:"canAddProCollectors"`
	CanStartProProcedures       bool      `json:"canStartProProcedures"`
	HasFullContentSearchAbility bool      `json:"hasFullContentSearchAbility"`
	HasOfflineModeAbility       bool      `json:"has_offline_mode_ability"`
	IsAssistanceEnabled         bool      `json:"is_assistance_enabled"`
}

type Capabilities struct {
	ShareSpaceStatus            string `json:"share_space_status"`
	Show2ddoc                   bool   `json:"show2ddoc"`
	HasOfflineModeAbility       bool   `json:"has_offline_mode_ability"`
	HasFullContentSearchAbility bool   `json:"hasFullContentSearchAbility"`
	SupportAvailable            bool   `json:"support_available"`
	CanAddProCollectors         bool   `json:"canAddProCollectors"`
	SecretQuestionAvailable     bool   `json:"secret_question_available"`
}

type Storage struct {
	SpaceUsed        int64 `json:"space_used"`
	SpaceFree        int64 `json:"space_free"`
	SpaceMax         int64 `json:"space_max"`
	SpaceNotComputed int64 `json:"space_not_computed"`
}

type Contracts struct {
	TOS struct {
		Version   string `json:"tos_version"`
		UpdatedAt string `json:"tos_updated_at"`
	} `json:",inline"`
	Offer struct {
		PID        string `json:"offer_pid"`
		UpdatedAt  string `json:"offer_updated_at"`
		NewOffer   bool   `json:"new_offer"`
		OtherOffer string `json:"other_offer"`
	} `json:",inline"`
	CCU struct {
		User   bool   `json:"ccu_user"`
		UserID string `json:"ccu_user_id"`
	} `json:",inline"`
}

type Profile struct {
	UserInfo     `json:",inline"`
	Offer        `json:"offer"`
	Capabilities `json:",inline"`
	Storage      `json:",inline"`
	Serializable `json:",inline"`

	Status                  string        `json:"status"`
	AuthorName              string        `json:"author_name"`
	LastConnexionDate       string        `json:"last_connexion_date"`
	VerifyProfile           string        `json:"verify_profile"`
	Completion              int           `json:"completion"`
	VerifiedDocuments       []interface{} `json:"verified_documents"`
	PartialAccount          bool          `json:"partial_account"`
	IDNumeriqueValid        bool          `json:"idn_valid"`
	BasicUser               bool          `json:"basic_user"`
	SecretQuestionAvailable bool          `json:"secret_question_available"`
	FirstConnection         bool          `json:"first_connection"`
	Salaried                bool          `json:"salaried"`
	IndexationConsent       bool          `json:"indexation_consent"`
}

//go:generate stringer -type=ProfileMode -linecomment

type ProfileMode int

const (
	ProfileModeDefault            ProfileMode = iota // default
	ProfileModeNoSpaceConsumption                    // without_space_consumption
)

// GetProfile returns the profile of the user.
func (c *Client) GetProfile(ctx context.Context, mode ProfileMode) (*Profile, error) { //nolint:interfacer
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
