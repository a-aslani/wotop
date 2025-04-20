package util

import (
	"strconv"
	"time"
)

// ParseUnixTimestamp converts a Unix timestamp to a *time.Time object.
//
// This function handles both second-based and millisecond-based Unix timestamps.
// If the input timestamp is 0, the function returns nil.
//
// Parameters:
//   - timestamp: The Unix timestamp to be parsed. It can be in seconds or milliseconds.
//
// Returns:
//   - A pointer to a time.Time object representing the parsed timestamp, or nil if the input is 0.
func ParseUnixTimestamp(timestamp int64) *time.Time {
	if timestamp == 0 {
		return nil
	}

	// Convert the timestamp to a string for length checking.
	timestampStr := strconv.FormatInt(timestamp, 10)

	// Check if the timestamp is at least 10 digits (seconds-based).
	if len(timestampStr) >= 10 {
		// If the timestamp is 13 digits (milliseconds-based), convert it to seconds.
		if len(timestampStr) == 13 {
			timestamp /= 1000
		}

		// Parse the timestamp into a time.Time object.
		parsedTime := time.Unix(timestamp, 0)
		return &parsedTime
	}

	// Return nil if the timestamp is invalid.
	return nil
}
