//go:build linux

package chrome

import cu "github.com/Davincible/chromedp-undetected"

func init() { //nolint:gochecknoinits
	chromeOpts = append(chromeOpts, cu.WithHeadless())
}
