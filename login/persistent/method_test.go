package persistent_test

import (
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/holyhope/digiposte-go-sdk/login"
	"github.com/holyhope/digiposte-go-sdk/login/persistent"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
)

var errBoom = errors.New("boom")

const testCookieName = "session"

func newTestCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:        testCookieName,
		Value:       value,
		Path:        "",
		Domain:      "",
		Expires:     time.Time{},
		RawExpires:  "",
		MaxAge:      0,
		Secure:      true,
		HttpOnly:    true,
		SameSite:    http.SameSiteLaxMode,
		Partitioned: false,
		Raw:         "",
		Unparsed:    nil,
		Quoted:      false,
	}
}

func newTestToken(accessToken string, expiry time.Time) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "",
		RefreshToken: "",
		Expiry:       expiry,
		ExpiresIn:    0,
	}
}

var _ = Describe("New", func() {
	var (
		store *fakeStore
		creds *login.Credentials

		inner        *mockedLoginMethod
		seedCookies  []*http.Cookie
		nbFactory    int
		factoryErr   error
		saveErrors   []error
		methodOption persistent.Option
	)

	BeforeEach(func() {
		store = &fakeStore{
			session: nil,
			loadErr: nil,
			saveErr: nil,
			nbLoads: 0,
			nbSaves: 0,
			saved:   nil,
		}
		creds = &login.Credentials{Username: "user", Password: "pass", OTPSecret: ""}
		inner = &mockedLoginMethod{
			Token:   newTestToken("fresh-token", time.Now().Add(time.Hour)),
			Cookies: []*http.Cookie{newTestCookie("fresh")},
			Err:     nil,
			nbCalls: 0,
		}
		seedCookies = nil
		nbFactory = 0
		factoryErr = nil
		saveErrors = nil
		methodOption = persistent.WithSaveErrorHandler(func(err error) {
			saveErrors = append(saveErrors, err)
		})
	})

	newMethod := func() persistent.NewMethod {
		return func(cookies []*http.Cookie) (login.Method, error) {
			nbFactory++
			seedCookies = cookies

			if factoryErr != nil {
				return nil, factoryErr
			}

			return inner, nil
		}
	}

	Context("When the store has no session", func() {
		It("Falls back to the factory with no seed cookies and saves the result", func() {
			m := persistent.New(store, newMethod(), methodOption)

			token, cookies, err := m.Login(context.Background(), creds)
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal(inner.Token))
			Expect(cookies).To(Equal(inner.Cookies))

			Expect(nbFactory).To(Equal(1))
			Expect(inner.nbCalls).To(Equal(1))
			Expect(seedCookies).To(BeEmpty())

			Expect(store.nbSaves).To(Equal(1))
			Expect(store.session).To(Equal(&persistent.Session{Token: inner.Token, Cookies: inner.Cookies}))
		})
	})

	Context("When the store returns a still-valid token", func() {
		BeforeEach(func() {
			store.session = &persistent.Session{
				Token:   newTestToken("stored-token", time.Now().Add(time.Hour)),
				Cookies: []*http.Cookie{newTestCookie("stored")},
			}
		})

		It("Resumes the stored session without calling the factory", func() {
			m := persistent.New(store, newMethod(), methodOption)

			token, cookies, err := m.Login(context.Background(), creds)
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal(store.session.Token))
			Expect(cookies).To(Equal(store.session.Cookies))

			Expect(nbFactory).To(Equal(0))
			Expect(inner.nbCalls).To(Equal(0))
		})

		It("Still saves the resumed session", func() {
			m := persistent.New(store, newMethod(), methodOption)

			_, _, err := m.Login(context.Background(), creds)
			Expect(err).ToNot(HaveOccurred())

			Expect(store.nbSaves).To(Equal(1))
		})
	})

	Context("When the store returns an expired token", func() {
		var storedCookies []*http.Cookie

		BeforeEach(func() {
			storedCookies = []*http.Cookie{newTestCookie("expired")}
			store.session = &persistent.Session{
				Token:   newTestToken("expired-token", time.Now().Add(-time.Hour)),
				Cookies: storedCookies,
			}
		})

		It("Falls back to the factory, seeding the stored cookies", func() {
			m := persistent.New(store, newMethod(), methodOption)

			token, _, err := m.Login(context.Background(), creds)
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal(inner.Token))

			Expect(nbFactory).To(Equal(1))
			Expect(seedCookies).To(Equal(storedCookies))
		})
	})

	Context("When the store returns an error while loading", func() {
		BeforeEach(func() {
			store.loadErr = errBoom
		})

		It("Falls back to the factory instead of failing the login", func() {
			m := persistent.New(store, newMethod(), methodOption)

			token, _, err := m.Login(context.Background(), creds)
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal(inner.Token))
			Expect(nbFactory).To(Equal(1))
		})
	})

	Context("When the store fails to save", func() {
		BeforeEach(func() {
			store.saveErr = errBoom
		})

		It("Still returns the obtained token and cookies successfully", func() {
			m := persistent.New(store, newMethod(), methodOption)

			token, cookies, err := m.Login(context.Background(), creds)
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal(inner.Token))
			Expect(cookies).To(Equal(inner.Cookies))
		})

		It("Reports the save error through the save-error handler", func() {
			m := persistent.New(store, newMethod(), methodOption)

			_, _, err := m.Login(context.Background(), creds)
			Expect(err).ToNot(HaveOccurred())

			Expect(saveErrors).To(HaveLen(1))
			Expect(saveErrors[0]).To(MatchError(errBoom))

			var saveErr *persistent.SaveError
			Expect(errors.As(saveErrors[0], &saveErr)).To(BeTrue())
		})
	})

	Context("When the factory returns an error", func() {
		BeforeEach(func() {
			factoryErr = errBoom
		})

		It("Returns the error and does not save anything", func() {
			m := persistent.New(store, newMethod(), methodOption)

			_, _, err := m.Login(context.Background(), creds)
			Expect(err).To(HaveOccurred())
			Expect(store.nbSaves).To(Equal(0))
		})
	})

	Context("When the inner login method returns an error", func() {
		BeforeEach(func() {
			inner.Err = errBoom
		})

		It("Returns the error and does not save anything", func() {
			m := persistent.New(store, newMethod(), methodOption)

			_, _, err := m.Login(context.Background(), creds)
			Expect(err).To(HaveOccurred())
			Expect(store.nbSaves).To(Equal(0))
		})
	})

	Context("When constructed with a nil store", func() {
		It("Returns an error instead of calling the factory", func() {
			m := persistent.New(nil, newMethod(), methodOption)

			_, _, err := m.Login(context.Background(), creds)
			Expect(err).To(HaveOccurred())
			Expect(nbFactory).To(Equal(0))
		})
	})

	Context("With no save-error handler configured", func() {
		BeforeEach(func() {
			store.saveErr = errBoom
		})

		It("Uses the default handler and still succeeds", func() {
			m := persistent.New(store, newMethod())

			_, _, err := m.Login(context.Background(), creds)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
