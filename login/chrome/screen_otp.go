package chrome

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type otpScreen struct {
	Secret string
}

var _ Screen = (*otpScreen)(nil)

func (s *otpScreen) String() string {
	return "OTP screen"
}

func (s *otpScreen) CurrentPageMatches(ctx context.Context) bool {
	var nodeIDs []cdp.NodeID

	if err := chromedp.Run(ctx,
		chromedp.NodeIDs(`#otpCode`, &nodeIDs, chromedp.ByID, chromedp.AtLeast(0)),
	); err != nil {
		errorLogger(ctx).Printf("run: %v\n", err)

		return false
	}

	return len(nodeIDs) > 0
}

var errEmptyOTP = fmt.Errorf("empty OTP secret")

func (s *otpScreen) Do(ctx context.Context) error {
	if s.Secret == "" {
		return errEmptyOTP
	}

	otpKey, err := otp.NewKeyFromURL(s.Secret)
	if err != nil {
		return fmt.Errorf("parse secret: %w", err)
	}

	otpCode, err := totp.GenerateCode(otpKey.Secret(), time.Now())
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	if err := (&chromedp.Tasks{
		// OTP not enabled, skip the screen
		chromedp.QueryAfter(`#linkLater`, func(ctx context.Context, eci runtime.ExecutionContextID, n ...*cdp.Node) error {
			if len(n) == 0 {
				return nil
			}

			if err := chromedp.MouseClickNode(n[0]).Do(ctx); err != nil {
				return fmt.Errorf("click: %w", err)
			}

			return nil
		}, chromedp.ByID, chromedp.AtLeast(0)),

		chromedp.WaitVisible(`#otpCode`, chromedp.ByID),
		chromedp.WaitEnabled(`#otpCode`, chromedp.ByID),
		chromedp.Clear(`#otpCode`, chromedp.ByID),
		chromedp.SendKeys(`#otpCode`, otpCode, chromedp.ByID),

		chromedp.WaitVisible(`#submit`, chromedp.ByID),
		chromedp.WaitEnabled(`#submit`, chromedp.ByID),
		chromedp.Click(`#submit`, chromedp.ByID),
	}).Do(ctx); err != nil {
		return fmt.Errorf("tasks: %w", err)
	}

	return nil
}

func (s *otpScreen) ShouldWaitForResponse() bool {
	return true
}

/*
func (c *chromeLogin) handleOTPError(ctx context.Context) error {
	var nodes []*cdp.Node

	if err := chromedp.Run(ctx,
		chromedp.Nodes(`.otp-error`, &nodes, chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	if len(nodes) == 0 {
		return nil
	}

	 errors.New(nodes[0].Children[0].NodeValue)
}
*/
