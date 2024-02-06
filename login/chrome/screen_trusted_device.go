package chrome

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type trustedDeviceScreen struct{}

var _ Screen = (*trustedDeviceScreen)(nil)

func (s *trustedDeviceScreen) String() string {
	return "trusted device screen"
}

func (s *trustedDeviceScreen) CurrentPageMatches(ctx context.Context) bool {
	var nodeIDs []cdp.NodeID

	err := chromedp.Run(ctx,
		chromedp.NodeIDs(`#save-trusted-device-form`, &nodeIDs, chromedp.ByID, chromedp.AtLeast(0)),
	)
	if err != nil {
		errorLogger(ctx).Printf("run: %v\n", err)

		return false
	}

	return len(nodeIDs) > 0
}

func (s *trustedDeviceScreen) Do(ctx context.Context) error {
	if err := (&chromedp.Tasks{
		chromedp.WaitVisible(`#linkLater`, chromedp.BySearch),
		chromedp.Click(`#linkLater`, chromedp.ByID),
	}).Do(ctx); err != nil {
		return fmt.Errorf("tasks: %w", err)
	}

	return nil
}

func (s *trustedDeviceScreen) ShouldWaitForResponse() bool {
	return true
}
