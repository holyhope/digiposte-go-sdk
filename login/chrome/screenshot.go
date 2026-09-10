package chrome

import (
	"context"
	"errors"
	"image/jpeg"

	"github.com/chromedp/chromedp"
)

func (c *chromeLogin) ScreenshotIfNeeded(ctx context.Context, err error) error {
	if !c.screenShortOnError {
		return err
	}

	return c.wrapWithScreenshot(ctx, err)
}

func (c *chromeLogin) wrapWithScreenshot(ctx context.Context, rootErr error) error {
	var imageData []byte

	err := chromedp.Run(ctx, chromedp.FullScreenshot(&imageData, jpeg.DefaultQuality))
	if err != nil {
		errorLogger(ctx).Printf("Failed to take screenshot: %v\n", err)

		return rootErr
	}

	infoLogger(ctx).Println("Screenshot taken")

	return &WithScreenshotError{
		Err:        rootErr,
		Screenshot: imageData,
	}
}

func GetScreenShot(err error) ([]byte, bool) {
	targetErr, ok := errors.AsType[*WithScreenshotError](err)
	if ok {
		return targetErr.Screenshot, true
	}

	return nil, false
}
