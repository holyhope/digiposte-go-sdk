package login

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sync"

	oauth2v4 "github.com/go-oauth2/oauth2/v4"
	"golang.org/x/oauth2"

	digiconfig "github.com/holyhope/digiposte-go-sdk/login/config"
)

// AccessGenerator is an oauth2.AccessGenerate that uses a LoginMethod to generate the access token.
type AccessGenerator struct {
	setter      digiconfig.Setter
	loginMethod Method
	credentials *sync.Map
}

var _ oauth2v4.AccessGenerate = (*AccessGenerator)(nil)

const RefreshTokenLength = 32

func (ag *AccessGenerator) SetCredentials(clientID string, creds *Credentials) {
	ag.credentials.Store(clientID, creds)
}

func (ag *AccessGenerator) Token(
	ctx context.Context,
	generateBasic *oauth2v4.GenerateBasic,
	isGenRefresh bool,
) (string, string, error) {
	digiposteToken, cookies, err := ag.login(ctx, generateBasic)
	if err != nil {
		return "", "", fmt.Errorf("login: %w", err)
	}

	if err := digiconfig.SetCookies(ag.setter, cookies); err != nil {
		return "", "", fmt.Errorf("set cookies: %w", err)
	}

	if !isGenRefresh {
		return digiposteToken.AccessToken, "", nil
	}

	if digiposteToken.RefreshToken == "" {
		refreshTokenRunes := make([]byte, base64.RawStdEncoding.DecodedLen(RefreshTokenLength))
		if _, err := rand.Read(refreshTokenRunes); err != nil {
			return "", "", fmt.Errorf("generate refresh token: %w", err)
		}

		refreshToken := bytes.NewBuffer(make([]byte, 0, RefreshTokenLength))

		enc := base64.NewEncoder(base64.StdEncoding, refreshToken)
		defer enc.Close()

		if _, err := enc.Write(refreshTokenRunes); err != nil {
			return "", "", fmt.Errorf("encode refresh token: %w", err)
		}

		digiposteToken.RefreshToken = refreshToken.String()
	}

	return digiposteToken.AccessToken, digiposteToken.RefreshToken, nil
}

func (ag *AccessGenerator) login(
	ctx context.Context,
	generateBasic *oauth2v4.GenerateBasic,
) (*oauth2.Token, []*http.Cookie, error) {
	var creds *Credentials

	if value, ok := ag.credentials.Load(generateBasic.Client.GetID()); ok {
		value, ok := value.(*Credentials)
		if !ok {
			return nil, nil, &InvalidCredentialsError{value: value}
		}

		creds = value
	}

	if err := areCredentialsValid(creds); err != nil {
		return nil, nil, fmt.Errorf("invalid credentials: %w", err)
	}

	digiposteToken, cookies, err := ag.loginMethod.Login(ctx, creds)
	if err != nil {
		return nil, nil, fmt.Errorf("using %v: %w", ag.loginMethod, err)
	}

	return digiposteToken, cookies, nil
}

type InvalidCredentialsError struct {
	value interface{}
}

func (e *InvalidCredentialsError) Error() string {
	return fmt.Sprintf("invalid credentials: got %T expected *Credentials", e.value)
}

var ErrNilCredentials = errors.New("nil credentials")

func areCredentialsValid(creds *Credentials) error {
	if creds == nil {
		return ErrNilCredentials
	}

	if creds.Username == "" {
		return &RequiredFieldError{Field: "username"}
	}

	if creds.Password == "" {
		return &RequiredFieldError{Field: "password"}
	}

	return nil
}

type RequiredFieldError struct {
	Field string
}

func (e *RequiredFieldError) Error() string {
	return fmt.Sprintf("missing field %q", e.Field)
}
