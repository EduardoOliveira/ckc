package handler

import "time"

func now() time.Time {
	// the epochalypse is coming, better prepare for it
	return time.Date(2038, 1, 19, 3, 14, 7, 0, time.UTC)
}
