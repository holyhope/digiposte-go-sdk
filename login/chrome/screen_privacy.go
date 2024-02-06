package chrome

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type privacyScreen struct {
	AcceptCookies bool
}

var _ Screen = (*privacyScreen)(nil)

func (s *privacyScreen) String() string {
	return "privacy screen"
}

func (s *privacyScreen) CurrentPageMatches(ctx context.Context) bool {
	var nodeIDs []cdp.NodeID

	if err := chromedp.Run(ctx,
		chromedp.NodeIDs(`#footer_tc_privacy_button_3`, &nodeIDs, chromedp.ByID, chromedp.AtLeast(0)),
	); err != nil {
		errorLogger(ctx).Printf("run: %v\n", err)

		return false
	}

	return len(nodeIDs) > 0
}

func (s *privacyScreen) Do(ctx context.Context) error {
	sel := `#footer_tc_privacy_button_3`

	if s.AcceptCookies {
		sel = `#footer_tc_privacy_button_2`
	}

	if err := chromedp.Click(sel, chromedp.ByID, chromedp.AtLeast(0)).Do(ctx); err != nil {
		return fmt.Errorf("click: %w", err)
	}

	return nil
}

func (s *privacyScreen) ShouldWaitForResponse() bool {
	return false
}
