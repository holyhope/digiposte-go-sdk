package chrome //nolint:testpackage

import "fmt"

func (e *WithScreenshotError) GomegaString() string {
	return fmt.Sprintf("screenshot: %v", byteCountSI(len(e.Screenshot)))
}

func byteCountSI(byteCount int) string {
	const unit = 1000
	if byteCount < unit {
		return fmt.Sprintf("%d B", byteCount)
	}

	div, exp := int64(unit), 0
	for n := byteCount / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB",
		float64(byteCount)/float64(div), "kMGTPE"[exp])
}
