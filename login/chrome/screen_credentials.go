package chrome

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

type credentialsScreen struct {
	Username string
	Password string
}

var _ Screen = (*credentialsScreen)(nil)

func (s *credentialsScreen) String() string {
	return "credentials screen"
}

func (s *credentialsScreen) CurrentPageMatches(ctx context.Context) bool {
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
	if err := (&chromedp.Tasks{
		chromedp.WaitVisible(`#submit`, chromedp.ByID),
		chromedp.WaitEnabled(`#submit`, chromedp.ByID),

		chromedp.WaitVisible(`#username`, chromedp.ByID),
		chromedp.Click(`#username`, chromedp.ByID),
		s.ClearInput(`#username`, chromedp.ByID),
		chromedp.SendKeys(`#username`, s.Username, chromedp.ByID),

		chromedp.WaitVisible(`#password`, chromedp.ByID),
		chromedp.Click(`#password`, chromedp.ByID),
		s.ClearInput(`#password`, chromedp.ByID),
		chromedp.SendKeys(`#password`, s.Password, chromedp.ByID),

		chromedp.Click(`#submit`, chromedp.ByID),
	}).Do(ctx); err != nil {
		return fmt.Errorf("tasks: %w", err)
	}

	return nil
}

func (s *credentialsScreen) ShouldWaitForResponse() bool {
	return true
}

func (s *credentialsScreen) ClearInput(sel interface{}, opts ...chromedp.QueryOption) *chromedp.Tasks {
	return &chromedp.Tasks{
		chromedp.Clear(sel, opts...),

		chromedp.SetValue(sel, "", opts...),

		input.DispatchKeyEvent(input.KeyDown).WithKey(kb.Control),
		input.DispatchKeyEvent(input.KeyDown).WithKey("a"),
		input.DispatchKeyEvent(input.KeyUp).WithKey("a"),
		input.DispatchKeyEvent(input.KeyUp).WithKey(kb.Control),
		input.DispatchKeyEvent(input.KeyDown).WithKey(kb.Backspace),
		input.DispatchKeyEvent(input.KeyUp).WithKey(kb.Backspace),
	}
}
