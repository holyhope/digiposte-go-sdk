package chrome

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

type credentialsScreen struct {
	Username string
	Password string

	// submitted is set once the credentials form has been submitted
	// successfully, so a slow navigation away from the login page (which can
	// take longer than a single resolver tick) doesn't cause the form to be
	// resubmitted with stale/duplicated input.
	submitted atomic.Bool
}

var _ Screen = (*credentialsScreen)(nil)

func (s *credentialsScreen) String() string {
	return "credentials screen"
}

func (s *credentialsScreen) CurrentPageMatches(ctx context.Context) bool {
	if s.submitted.Load() {
		return false
	}

	if s.Username == "" || s.Password == "" {
		return false
	}

	var nodeIDs []cdp.NodeID

	err := chromedp.Run(ctx,
		chromedp.NodeIDs(`form[name=login-form]`, &nodeIDs, chromedp.ByQuery, chromedp.AtLeast(0)),
	)
	if err != nil {
		errorLogger(ctx).Printf("run: %v\n", err)

		return false
	}

	return len(nodeIDs) > 0
}

func (s *credentialsScreen) Do(ctx context.Context) error {
	err := (&chromedp.Tasks{
		chromedp.WaitVisible(`#submit-button`, chromedp.ByID),
		chromedp.WaitEnabled(`#submit-button`, chromedp.ByID),

		chromedp.WaitVisible(`#username`, chromedp.ByID),
		chromedp.Click(`#username`, chromedp.ByID),
		s.ClearInput(`#username`, chromedp.ByID),
		chromedp.SendKeys(`#username`, s.Username, chromedp.ByID),

		chromedp.WaitVisible(`#password`, chromedp.ByID),
		chromedp.Click(`#password`, chromedp.ByID),
		s.ClearInput(`#password`, chromedp.ByID),
		chromedp.SendKeys(`#password`, s.Password, chromedp.ByID),

		chromedp.Click(`#submit-button`, chromedp.ByID),
	}).Do(ctx)
	if err != nil {
		return fmt.Errorf("tasks: %w", err)
	}

	s.submitted.Store(true)

	return nil
}

func (s *credentialsScreen) ShouldWaitForResponse() bool {
	return true
}

func (s *credentialsScreen) ClearInput(sel any, opts ...chromedp.QueryOption) *chromedp.Tasks {
	return &chromedp.Tasks{
		chromedp.Clear(sel, opts...),

		// Note: intentionally not using chromedp.SetValue here. It dispatches
		// synthetic "input"/"change" events and then verifies the field still
		// holds the value it just set; some login pages mutate the field's
		// value synchronously in response to those events (client-side
		// validation/formatting), which makes that verification flaky even
		// though the field is genuinely empty. The key events below clear it
		// the same way a real user would.

		input.DispatchKeyEvent(input.KeyDown).WithKey(kb.Control),
		input.DispatchKeyEvent(input.KeyDown).WithKey("a"),
		input.DispatchKeyEvent(input.KeyUp).WithKey("a"),
		input.DispatchKeyEvent(input.KeyUp).WithKey(kb.Control),
		input.DispatchKeyEvent(input.KeyDown).WithKey(kb.Backspace),
		input.DispatchKeyEvent(input.KeyUp).WithKey(kb.Backspace),
	}
}
