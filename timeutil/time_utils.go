package timeutil

import (
	"fmt"
	"time"
)

// UTCNow returns the current time in UTC timezone.
// This should be used instead of time.Now() throughout the application
// to ensure all database timestamps are stored in UTC.
func UTCNow() time.Time {
	return time.Now().UTC()
}

// UTCTime converts a given time to UTC timezone.
// Use this when you need to ensure a time value is in UTC.
func UTCTime(t time.Time) time.Time {
	return t.UTC()
}

// ParseTimeToUTC parses a time string and converts it to UTC.
// The layout parameter should follow Go's time parsing format.
func ParseTimeToUTC(layout, value string) (time.Time, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// ParseTimeInLocationToUTC parses a time string in a specific location and converts it to UTC.
func ParseTimeInLocationToUTC(layout, value string, loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation(layout, value, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// AddDurationUTC adds a duration to the current UTC time.
// This is a convenience function for calculating future UTC times.
func AddDurationUTC(duration time.Duration) time.Time {
	return UTCNow().Add(duration)
}

// FormatUTCTime formats a UTC time for consistent string representation.
// Uses RFC3339 format which includes timezone information.
func FormatUTCTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// ConvertToUserTimezone converts a UTC time to the user's timezone.
// Use this when displaying times to users.
func ConvertToUserTimezone(utcTime time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return utcTime.In(loc), nil
}

// TimeValueOrZeroUTC safely dereferences a time pointer and ensures it's in UTC,
// returning the UTC time value or zero time if nil.
func TimeValueOrZeroUTC(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.UTC()
}

// TimePointerOrNilUTC converts a time value to pointer and ensures it's in UTC,
// returning nil if the time is zero.
func TimePointerOrNilUTC(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	utcTime := t.UTC()
	return &utcTime
}

// IsExpiredUTC checks if a UTC time has passed relative to current UTC time.
func IsExpiredUTC(expiryTime time.Time) bool {
	return expiryTime.Before(UTCNow())
}

// DurationUntilExpiry returns the duration until the given UTC time expires.
// Returns 0 if already expired.
func DurationUntilExpiry(expiryTime time.Time) time.Duration {
	duration := expiryTime.Sub(UTCNow())
	if duration < 0 {
		return 0
	}
	return duration
}

// FormatDurationHoursMinutes converts minutes to "Xh Ymin" format.
// Examples: 0 -> "0min", 45 -> "45min", 60 -> "1h", 153 -> "2h 33min"
// This is useful for displaying study time, video duration, etc.
func FormatDurationHoursMinutes(minutes int) string {
	if minutes == 0 {
		return "0min"
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60

	if hours == 0 {
		return intToString(remainingMinutes) + "min"
	}

	if remainingMinutes == 0 {
		return intToString(hours) + "h"
	}

	return intToString(hours) + "h " + intToString(remainingMinutes) + "min"
}

// intToString converts an integer to string. Helper to avoid importing strconv.
func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	result := ""
	for n > 0 {
		result = string(rune('0'+(n%10))) + result
		n /= 10
	}

	if negative {
		result = "-" + result
	}

	return result
}

// FormatDuration converts total seconds to a human-readable duration string.
// Examples: 45 -> "45s", 120 -> "2min", 150 -> "2min 30s", 3660 -> "1h 1min"
func FormatDuration(totalSeconds int) string {
	if totalSeconds < 60 {
		return fmt.Sprintf("%ds", totalSeconds)
	}
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	if minutes < 60 {
		if seconds > 0 {
			return fmt.Sprintf("%dmin %ds", minutes, seconds)
		}
		return fmt.Sprintf("%dmin", minutes)
	}
	hours := minutes / 60
	remainingMinutes := minutes % 60
	if remainingMinutes > 0 {
		return fmt.Sprintf("%dh %dmin", hours, remainingMinutes)
	}
	return fmt.Sprintf("%dh", hours)
}
