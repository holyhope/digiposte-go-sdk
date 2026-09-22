package chrome

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/onsi/gomega/gbytes"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
)

// slowTestScreen is a Screen whose Do() takes longer than a short polling
// frequency but less than a generous screen-execution timeout. It never
// touches chromedp itself; resolveScreen is swapped in the spec below to
// call Do() directly, since chromedp.Run requires a live browser and this
// spec only needs to exercise Screens.run's timeout-selection logic.
type slowTestScreen struct {
	sleep time.Duration
	calls atomic.Int32
}

var _ Screen = (*slowTestScreen)(nil)

func (s *slowTestScreen) String() string { return "slow test screen" }

func (s *slowTestScreen) CurrentPageMatches(_ context.Context) bool { return true }

func (s *slowTestScreen) ShouldWaitForResponse() bool { return false }

func (s *slowTestScreen) Do(ctx context.Context) error {
	s.calls.Add(1)

	select {
	case <-time.After(s.sleep):
		return nil
	case <-ctx.Done():
		return fmt.Errorf("wait: %w", ctx.Err())
	}
}

var _ = Describe("Screens", func() {
	Describe("run", func() {
		It("gives a screen the full execution timeout despite a shorter polling frequency", func(ctx SpecContext) {
			// Bypass chromedp (which requires a live browser) so this spec can
			// run screen.Do() directly and observe the deadline Screens.run
			// actually applies.
			original := resolveScreen
			resolveScreen = func(ctx context.Context, screen Screen) error {
				return screen.Do(ctx)
			}

			defer func() { resolveScreen = original }()

			screen := &slowTestScreen{
				sleep: 150 * time.Millisecond,
				calls: atomic.Int32{},
			}

			screens := &Screens{
				screens:          nil,
				refreshFrequency: 20 * time.Millisecond,
				screenTimeout:    time.Second,
				succeeded:        atomic.Bool{},
			}

			logs := gbytes.NewBuffer()

			runCtx := withErrorLogger(
				withInfoLogger(ctx, log.New(logs, "[INFO] ", 0)),
				log.New(logs, "[ERRO] ", 0),
			)

			runCtx, cancel := context.WithCancel(runCtx)
			defer cancel()

			done := make(chan struct{})

			go func() {
				defer close(done)

				screens.run(runCtx, screen)
			}()

			// The polling frequency (20ms) is well below the screen's Do()
			// duration (150ms). If the per-attempt timeout were still derived
			// from refreshFrequency (the bug this change fixes), Do() would
			// always be killed by context.DeadlineExceeded before it could
			// return, and "Screen passed" would never appear.
			Eventually(logs).Should(gbytes.Say("Screen passed"))

			cancel()

			Eventually(done).Should(BeClosed())

			Expect(screen.calls.Load()).To(BeNumerically(">=", 1))
			Expect(logs.Contents()).ToNot(ContainSubstring("context deadline exceeded"))
		})
	})
})
