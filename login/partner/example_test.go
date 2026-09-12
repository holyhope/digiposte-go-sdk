package partner_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/holyhope/digiposte-go-sdk/login/partner"
)

// ExampleNewClientCredentialsSource demonstrates the client_credentials
// grant end-to-end. In a real program, Endpoint and OkapiKey come from your
// own Partner API OKAPI/Swagger contract, not a local test server.
func ExampleNewClientCredentialsSource() {
	// A stub Partner API token endpoint, standing in for
	// https://api.laposte.fr/digiposte/v3/token (or your own OKAPI space's
	// equivalent).
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.Header.Get(partner.OkapiKeyHeader) == "" {
			http.Error(resp, "missing Okapi key", http.StatusUnauthorized)

			return
		}

		resp.Header().Set("Content-Type", "application/json")
		_, _ = resp.Write([]byte(`{"access_token":"example-access-token","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	source, err := partner.NewClientCredentialsSource(context.Background(), partner.ClientCredentialsConfig{ //nolint:gosec
		ClientID:     "example-client-id",
		ClientSecret: "example-client-secret",
		Endpoint: partner.Endpoint{
			AuthURL:   "",
			TokenURL:  server.URL,
			AuthStyle: 0,
		},
		OkapiKey: "example-okapi-key",
		Scopes:   nil,
		Timeout:  0,
	})
	if err != nil {
		panic(fmt.Errorf("new client credentials source: %w", err))
	}

	token, err := source.Token()
	if err != nil {
		panic(fmt.Errorf("get token: %w", err))
	}

	fmt.Printf("Token type: %s\n", token.Type())
	fmt.Printf("Token valid: %v\n", token.Valid())

	// Output:
	// Token type: Bearer
	// Token valid: true
}
