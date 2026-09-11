package persistent_test

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login"
)

// mockedLoginMethod is a mock of login.Method.
type mockedLoginMethod struct {
	Token   *oauth2.Token
	Cookies []*http.Cookie
	Err     error

	nbCalls int

	// lastSeedCookies is not used by the login.Method interface itself; it
	// is only used to assert what cookies a factory built this mock with,
	// via a closure in the test that constructed it.
}

func (m *mockedLoginMethod) Login(_ context.Context, _ *login.Credentials) (*oauth2.Token, []*http.Cookie, error) {
	m.nbCalls++

	if m.Err != nil {
		return nil, nil, m.Err
	}

	return m.Token, m.Cookies, nil
}
