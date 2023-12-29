package utils

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func UnixString2Time(unix string) (time.Time, error) {
	unixFloat, err := strconv.ParseFloat(unix, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse access_expires_at: %w", err)
	}

	return UnixFloat2Time(unixFloat), nil
}

func UnixFloat2Time(unix float64) time.Time {
	sec := math.Trunc(unix)
	nano := (unix - sec) * float64(time.Second/time.Nanosecond)

	return time.Unix(int64(sec), int64(nano))
}
