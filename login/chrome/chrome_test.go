package chrome_test

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/go-rod/rod/lib/launcher"

	"github.com/holyhope/digiposte-go-sdk/login"
	"github.com/holyhope/digiposte-go-sdk/login/chrome"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
)

var _ = Describe("Login", func() {
	Context("With valid options", func() {
		var username, password, otpSecret string

		var debugScreenshot []byte

		var chromeMethod login.Method

		BeforeEach(func() {
			username = os.Getenv("DIGIPOSTE_USERNAME")
			if len(username) == 0 {
				Skip("missing DIGIPOSTE_USERNAME")
			}

			password = os.Getenv("DIGIPOSTE_PASSWORD")
			if len(password) == 0 {
				Skip("missing DIGIPOSTE_PASSWORD")
			}

			otpSecret = os.Getenv("DIGIPOSTE_OTP_SECRET")

			chromeBinary, err := launcher.NewBrowser().Get()
			Expect(err).ToNot(HaveOccurred())

			loginWithChrome, err := chrome.New(
				chrome.WithURL(os.Getenv("DIGIPOSTE_URL")),
				chrome.WithCookies(nil),
				chrome.WithRefreshFrequency(500*time.Millisecond), // Reduce the test duration
				chrome.WithLoggers(
					log.New(GinkgoWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lmsgprefix),
					log.New(GinkgoWriter, "[ERRO] ", log.Ldate|log.Ltime|log.Lmsgprefix),
				),
				chrome.WithScreenShortOnError(),
				chrome.WithTimeout(3*time.Minute),
				chrome.WithBinary(chromeBinary),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(loginWithChrome).ToNot(BeNil())

			chromeMethod = loginWithChrome
		})

		AfterEach(func() {
			if len(debugScreenshot) == 0 {
				return
			}

			cwd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			screenshotPath := path.Join(cwd, CurrentSpecReport().FullText()+".png")
			defer Expect(os.WriteFile(
				screenshotPath, // Use GinkgoT().TempDir() instead?
				debugScreenshot,
				0o600,
			)).To(Succeed())

			fmt.Fprintf(GinkgoWriter, "Screenshot saved to %q\n", screenshotPath)
		})

		It("Should work", func(ctx SpecContext) {
			token, cookies, err := chromeMethod.Login(
				ctx,
				&login.Credentials{
					Username:  username,
					Password:  password,
					OTPSecret: otpSecret,
				},
			)
			if err != nil {
				if screenshot, ok := chrome.GetScreenShot(err); ok {
					debugScreenshot = screenshot
				}
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(token.Valid()).To(BeTrue())
			Expect(cookies).ToNot(BeEmpty())

			fmt.Fprintf(GinkgoWriter, "Token expires at %v\n", token.Expiry.Local())
		})
	})

	Context("With invalid options", func() {
		Describe("Empty URL", func() {
			It("Should return an error", func() {
				_, err := chrome.New(
					chrome.WithURL(""),
				)
				Expect(err).To(MatchError(HaveSuffix(`option "WithURL": url is empty`)))
			})
		})

		Describe("Negative refresh frequency", func() {
			It("Should return an error", func() {
				_, err := chrome.New(
					chrome.WithRefreshFrequency(-1),
				)
				Expect(err).To(MatchError(HaveSuffix(`option "WithRefreshFrequency": frequency must be positive`)))
			})
		})

		Describe("Negative timeout", func() {
			It("Should return an error", func() {
				_, err := chrome.New(
					chrome.WithTimeout(-1),
				)
				Expect(err).To(MatchError(HaveSuffix(`option "WithTimeout": timeout must be positive`)))
			})
		})
	})
})
