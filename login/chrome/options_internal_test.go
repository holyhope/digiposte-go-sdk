package chrome

import (
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
)

// Applying WithScreenTimeout requires access to chromeLogin's unexported
// screenTimeout field, so this spec lives in the internal (white-box)
// package rather than chrome_test alongside the rejection specs.
var _ = Describe("WithScreenTimeout", func() {
	It("sets the chromeLogin's screen execution timeout, independently of refreshFrequency", func() {
		instance := &chromeLogin{
			url:                "",
			cookies:            nil,
			screenShortOnError: false,
			refreshFrequency:   500 * time.Millisecond,
			screenTimeout:      0,
			timeout:            0,
			infoLogger:         nil,
			errorLogger:        nil,
			binaryPath:         "",
		}

		Expect(WithScreenTimeout(5 * time.Second).Apply(instance)).To(Succeed())

		Expect(instance.screenTimeout).To(Equal(5 * time.Second))
		Expect(instance.refreshFrequency).To(Equal(500 * time.Millisecond))
	})
})
