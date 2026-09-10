package login

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// Credentials is the credentials to connect to digiposte.
type Credentials struct {
	Username  string
	Password  string
	OTPSecret string
}

type Option interface {
	Apply(instance any) error
}

type OptionFunc func(instance any) error

func (f OptionFunc) Apply(instance any) error {
	return f(instance)
}

// Method is the method to connect to digiposte.
type Method interface {
	Login(ctx context.Context, creds *Credentials) (*oauth2.Token, []*http.Cookie, error)
}

type MethodFunc func(ctx context.Context, creds *Credentials) (*oauth2.Token, []*http.Cookie, error)

func (f MethodFunc) Login(ctx context.Context, creds *Credentials) (*oauth2.Token, []*http.Cookie, error) {
	return f(ctx, creds)
}

type InvalidOptionError struct {
	Name string
	Err  error
}

func (e *InvalidOptionError) Error() string {
	return fmt.Sprintf("option %q: %v", e.Name, e.Err)
}

func (e *InvalidOptionError) Unwrap() error {
	return e.Err
}

func (e *InvalidOptionError) Apply(any) error {
	return e
}
