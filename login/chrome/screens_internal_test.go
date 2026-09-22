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

// directResolver is a fake screenResolver that calls screen.Do() directly,
// bypassing chromedp (which requires a live browser). It lets this spec
// exercise Screens.run's per-attempt timeout selection in isolation.
type directResolver struct{}

func (directResolver) resolve(ctx context.Context, screen Screen) error {
	err := screen.Do(ctx)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}

	return nil
}

// slowTestScreen is a Screen whose Do() takes longer than a short polling
// frequency but less than a generous screen-execution timeout.
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
			screen := &slowTestScreen{
				sleep: 150 * time.Millisecond,
				calls: atomic.Int32{},
			}

			screens := &Screens{
				screens:          nil,
				refreshFrequency: 20 * time.Millisecond,
				screenTimeout:    time.Second,
				resolver:         directResolver{},
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
