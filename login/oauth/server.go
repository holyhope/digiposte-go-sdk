package oauth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	oautherrs "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"

	"github.com/holyhope/digiposte-go-sdk/login"
	digiconfig "github.com/holyhope/digiposte-go-sdk/login/config"
)

const (
	// AuthorizePath is the path to the authorize endpoint.
	AuthorizePath = "/authorize"
	// TokenPath is the path to the token endpoint.
	TokenPath = "/token"

	// ReadTimeout is the timeout for reading the request.
	ReadTimeout = 5 * time.Second
	// WriteTimeout is the timeout for writing the response.
	WriteTimeout = 5 * time.Minute
)

// Server is a local web server for collecting auth.
type Server struct {
	server          *http.Server
	listener        net.Listener
	manager         *manage.Manager
	clientStore     *store.ClientStore
	accessGenerator *AccessGenerator
}

type Config struct {
	Addr        string
	Server      *server.Config
	LoginMethod login.Method
	Logger      *log.Logger
}

// StartServer starts a local webserver to receive the auth.
func NewServer(setter digiconfig.Setter, config *Config) (*Server, error) {
	// client memory store
	clientStore := store.NewClientStore()

	accessGenerator := &AccessGenerator{
		setter:      setter,
		loginMethod: config.LoginMethod,
		credentials: &sync.Map{},
	}

	manager := newManager(clientStore, accessGenerator)

	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	return &Server{
		server:          newServer(manager, config.Server, listener, config.Logger),
		listener:        listener,
		manager:         manager,
		clientStore:     clientStore,
		accessGenerator: accessGenerator,
	}, nil
}

// RegisterUser adds a user to the server.
func (s *Server) RegisterUser(clientID, clientSecret, redirectURL, username, password, otpSecret string) error {
	if err := s.clientStore.Set(clientID, &models.Client{
		ID:     clientID,
		Secret: clientSecret,
		UserID: clientID,
		Public: false,
		Domain: redirectURL,
	}); err != nil {
		return fmt.Errorf("set client: %w", err)
	}

	s.accessGenerator.SetCredentials(clientID, &login.Credentials{
		Username:  username,
		Password:  password,
		OTPSecret: otpSecret,
	})

	return nil
}

func newServer(manager oauth2.Manager, config *server.Config, listener net.Listener, logger *log.Logger) *http.Server {
	mux := http.NewServeMux()
	oauthServer := server.NewServer(config, manager)

	mux.HandleFunc(AuthorizePath, func(w http.ResponseWriter, r *http.Request) {
		if err := oauthServer.HandleAuthorizeRequest(w, r); err != nil {
			logger.Printf("Failed to handle authorize request: %v", err)
		}
	})
	mux.HandleFunc(TokenPath, func(w http.ResponseWriter, r *http.Request) {
		if err := oauthServer.HandleTokenRequest(w, r); err != nil {
			logger.Printf("Failed to handle token request: %v", err)
		}
	})

	oauthServer.SetAllowGetAccessRequest(true)

	oauthServer.UserAuthorizationHandler = func(w http.ResponseWriter, r *http.Request) (string, error) {
		if err := r.ParseForm(); err != nil {
			return "", oautherrs.ErrInvalidRequest
		}

		return r.Form.Get("client_id"), nil
	}

	oauthServer.ClientInfoHandler = func(r *http.Request) (string, string, error) {
		if err := r.ParseForm(); err != nil {
			return "", "", oautherrs.ErrInvalidRequest
		}

		return r.Form.Get("client_id"), r.Form.Get("client_secret"), nil
	}

	httpServer := &http.Server{
		Addr:              listener.Addr().String(),
		ErrorLog:          logger,
		Handler:           mux,
		ReadHeaderTimeout: ReadTimeout,
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       WriteTimeout,

		TLSConfig:      nil,
		MaxHeaderBytes: 0,
		BaseContext:    nil,
		TLSNextProto:   nil,
		ConnState:      nil,
		ConnContext:    nil,
	}

	return httpServer
}

func newManager(cs oauth2.ClientStore, ag oauth2.AccessGenerate) *manage.Manager {
	manager := manage.NewDefaultManager()

	manager.MustTokenStorage(store.NewMemoryTokenStore())
	manager.MapClientStorage(cs)
	manager.MapAccessGenerate(ag)
	manager.SetRefreshTokenCfg(&manage.RefreshingConfig{
		AccessTokenExp:     time.Hour,
		IsGenerateRefresh:  true,
		RefreshTokenExp:    0,
		IsResetRefreshTime: false,
		IsRemoveAccess:     false,
		IsRemoveRefreshing: true,
	})

	return manager
}

// Close closes the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx) //nolint:wrapcheck
}

// Start starts the server.
func (s *Server) Start() error {
	if err := s.server.Serve(s.listener); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

// AuthorizeURL returns the URL to the authorize endpoint.
func (s *Server) AuthorizeURL() string {
	return "http://" + s.listener.Addr().String() + AuthorizePath
}

// TokenURL returns the URL to the token endpoint.
func (s *Server) TokenURL() string {
	return "http://" + s.listener.Addr().String() + TokenPath
}
