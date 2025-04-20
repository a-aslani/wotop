package wotop_util

import (
	"strconv"
	"time"
)

func ParseUnixTimestamp(timestamp int64) *time.Time {
	if timestamp == 0 {
		return nil
	}

	timestampStr := strconv.FormatInt(timestamp, 10)

	if len(timestampStr) >= 10 {
		if len(timestampStr) == 13 {
			timestamp /= 1000
		}

		parsedTime := time.Unix(timestamp, 0)
		return &parsedTime
	}

	return nil
}
