package utils

import (
	"context"
	"fmt"

	"github.com/go-rod/rod/lib/launcher"
)

func GetChrome(ctx context.Context) (string, error) {
	browser := launcher.NewBrowser()

	browser.Context = ctx

	path, err := browser.Get()
	if err != nil {
		return "", fmt.Errorf("browser download: %w", err)
	}

	if err := browser.Validate(); err != nil {
		return "", fmt.Errorf("browser download validation: %w", err)
	}

	return path, nil
}
