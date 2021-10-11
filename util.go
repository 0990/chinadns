package chinadns

import "time"

func timeSinceMS(t time.Time) int64 {
	return int64(time.Since(t) / 1e6)
}
