package persistent_test

import (
	"context"

	"github.com/holyhope/digiposte-go-sdk/login/persistent"
)

// fakeStore is a minimal in-memory persistent.Store for tests only. It is
// intentionally not exported by the persistent package: the package ships
// no default Store implementation.
type fakeStore struct {
	session *persistent.Session

	loadErr error
	saveErr error

	nbLoads int
	nbSaves int

	saved []*persistent.Session
}

func (s *fakeStore) Load(_ context.Context) (*persistent.Session, error) {
	s.nbLoads++

	if s.loadErr != nil {
		return nil, s.loadErr
	}

	if s.session == nil {
		return nil, persistent.ErrSessionNotFound
	}

	return s.session, nil
}

func (s *fakeStore) Save(_ context.Context, session *persistent.Session) error {
	s.nbSaves++

	if s.saveErr != nil {
		return s.saveErr
	}

	s.session = session
	s.saved = append(s.saved, session)

	return nil
}
