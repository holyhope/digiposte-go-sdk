package persistent_test

import (
	"context"
	"time"

	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login/persistent"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
)

var _ = Describe("Store contract", func() {
	var store *fakeStore

	BeforeEach(func() {
		store = &fakeStore{
			session: nil,
			loadErr: nil,
			saveErr: nil,
			nbLoads: 0,
			nbSaves: 0,
			saved:   nil,
		}
	})

	Context("When no session has been saved yet", func() {
		It("Reports ErrSessionNotFound distinctly from a generic error", func() {
			session, err := store.Load(context.Background())
			Expect(session).To(BeNil())
			Expect(err).To(MatchError(persistent.ErrSessionNotFound))
		})
	})

	Context("When a session has been saved", func() {
		It("Round-trips the saved session", func() {
			saved := &persistent.Session{
				Token: &oauth2.Token{
					AccessToken:  "at",
					TokenType:    "",
					RefreshToken: "",
					Expiry:       time.Time{},
					ExpiresIn:    0,
				},
				Cookies: nil,
			}

			Expect(store.Save(context.Background(), saved)).To(Succeed())

			loaded, err := store.Load(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(loaded).To(Equal(saved))
		})
	})
})
