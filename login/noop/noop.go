package noop

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login"
)

type LoginMethod struct {
	Token   *oauth2.Token
	Cookies []*http.Cookie
}

func (lm *LoginMethod) Login(_ context.Context, _ *login.Credentials) (*oauth2.Token, []*http.Cookie, error) {
	return lm.Token, lm.Cookies, nil
}
