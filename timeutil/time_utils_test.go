package timeutil

import (
	"testing"
	"time"
)

func TestUTCNow(t *testing.T) {
	now := UTCNow()

	// Verify it's in UTC
	if now.Location() != time.UTC {
		t.Errorf("UTCNow() returned time in %v, expected UTC", now.Location())
	}

	// Verify it's recent (within last second)
	if time.Since(now) > time.Second {
		t.Errorf("UTCNow() returned time that's not recent: %v", now)
	}
}

func TestUTCTime(t *testing.T) {
	// Test with time in different timezone
	est, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2024, 1, 1, 12, 0, 0, 0, est)

	utcTime := UTCTime(localTime)

	// Should be in UTC
	if utcTime.Location() != time.UTC {
		t.Errorf("UTCTime() returned time in %v, expected UTC", utcTime.Location())
	}

	// Should represent the same moment
	if !localTime.Equal(utcTime) {
		t.Errorf("UTCTime() changed the time moment: %v != %v", localTime, utcTime)
	}
}

func TestParseTimeToUTC(t *testing.T) {
	layout := "2006-01-02 15:04:05"
	value := "2024-01-01 12:00:00"

	parsedTime, err := ParseTimeToUTC(layout, value)
	if err != nil {
		t.Fatalf("ParseTimeToUTC() returned error: %v", err)
	}

	if parsedTime.Location() != time.UTC {
		t.Errorf("ParseTimeToUTC() returned time in %v, expected UTC", parsedTime.Location())
	}
}

func TestParseTimeInLocationToUTC(t *testing.T) {
	layout := "2006-01-02 15:04:05"
	value := "2024-01-01 12:00:00"
	est, _ := time.LoadLocation("America/New_York")

	parsedTime, err := ParseTimeInLocationToUTC(layout, value, est)
	if err != nil {
		t.Fatalf("ParseTimeInLocationToUTC() returned error: %v", err)
	}

	if parsedTime.Location() != time.UTC {
		t.Errorf("ParseTimeInLocationToUTC() returned time in %v, expected UTC", parsedTime.Location())
	}
}

func TestAddDurationUTC(t *testing.T) {
	duration := 1 * time.Hour
	futureTime := AddDurationUTC(duration)

	if futureTime.Location() != time.UTC {
		t.Errorf("AddDurationUTC() returned time in %v, expected UTC", futureTime.Location())
	}

	// Should be approximately 1 hour from now
	expectedTime := UTCNow().Add(duration)
	diff := futureTime.Sub(expectedTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("AddDurationUTC() returned unexpected time, diff: %v", diff)
	}
}

func TestFormatUTCTime(t *testing.T) {
	utcTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	formatted := FormatUTCTime(utcTime)

	expected := "2024-01-01T12:00:00Z"
	if formatted != expected {
		t.Errorf("FormatUTCTime() = %v, expected %v", formatted, expected)
	}
}

func TestConvertToUserTimezone(t *testing.T) {
	utcTime := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC) // 5 PM UTC

	userTime, err := ConvertToUserTimezone(utcTime, "America/New_York")
	if err != nil {
		t.Fatalf("ConvertToUserTimezone() returned error: %v", err)
	}

	// In January, EST is UTC-5, so 5 PM UTC should be 12 PM EST
	if userTime.Hour() != 12 {
		t.Errorf("ConvertToUserTimezone() hour = %v, expected 12", userTime.Hour())
	}
}

func TestTimeValueOrZeroUTC(t *testing.T) {
	// Test with nil pointer
	result := TimeValueOrZeroUTC(nil)
	if !result.IsZero() {
		t.Errorf("TimeValueOrZeroUTC(nil) should return zero time")
	}

	// Test with valid time pointer in different timezone
	est, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2024, 1, 1, 12, 0, 0, 0, est)

	result = TimeValueOrZeroUTC(&localTime)
	if result.Location() != time.UTC {
		t.Errorf("TimeValueOrZeroUTC() returned time in %v, expected UTC", result.Location())
	}

	if !result.Equal(localTime) {
		t.Errorf("TimeValueOrZeroUTC() changed the time moment")
	}
}

func TestTimePointerOrNilUTC(t *testing.T) {
	// Test with zero time
	zeroTime := time.Time{}
	result := TimePointerOrNilUTC(zeroTime)
	if result != nil {
		t.Errorf("TimePointerOrNilUTC(zero) should return nil")
	}

	// Test with valid time in different timezone
	est, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2024, 1, 1, 12, 0, 0, 0, est)

	result = TimePointerOrNilUTC(localTime)
	if result == nil {
		t.Errorf("TimePointerOrNilUTC() should not return nil for valid time")
		return
	}

	if result.Location() != time.UTC {
		t.Errorf("TimePointerOrNilUTC() returned time in %v, expected UTC", result.Location())
	}

	if !result.Equal(localTime) {
		t.Errorf("TimePointerOrNilUTC() changed the time moment")
	}
}

func TestIsExpiredUTC(t *testing.T) {
	// Test with past time
	pastTime := UTCNow().Add(-1 * time.Hour)
	if !IsExpiredUTC(pastTime) {
		t.Errorf("IsExpiredUTC() should return true for past time")
	}

	// Test with future time
	futureTime := UTCNow().Add(1 * time.Hour)
	if IsExpiredUTC(futureTime) {
		t.Errorf("IsExpiredUTC() should return false for future time")
	}
}

func TestDurationUntilExpiry(t *testing.T) {
	now := UTCNow()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	// Test future time
	duration := DurationUntilExpiry(future)
	if duration <= 0 {
		t.Errorf("DurationUntilExpiry() for future time should be positive, got %v", duration)
	}

	// Test past time
	duration = DurationUntilExpiry(past)
	if duration != 0 {
		t.Errorf("DurationUntilExpiry() should return 0 for past time, got %v", duration)
	}
}

func TestFormatDurationHoursMinutes(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		expected string
	}{
		{
			name:     "zero minutes",
			minutes:  0,
			expected: "0min",
		},
		{
			name:     "only minutes",
			minutes:  45,
			expected: "45min",
		},
		{
			name:     "exactly one hour",
			minutes:  60,
			expected: "1h",
		},
		{
			name:     "hours and minutes",
			minutes:  153,
			expected: "2h 33min",
		},
		{
			name:     "multiple hours",
			minutes:  180,
			expected: "3h",
		},
		{
			name:     "single minute",
			minutes:  1,
			expected: "1min",
		},
		{
			name:     "large duration",
			minutes:  725,
			expected: "12h 5min",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDurationHoursMinutes(tt.minutes)
			if result != tt.expected {
				t.Errorf("FormatDurationHoursMinutes(%d) = %q, expected %q", tt.minutes, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{
			name:     "less than a minute",
			seconds:  45,
			expected: "45s",
		},
		{
			name:     "exactly one minute",
			seconds:  60,
			expected: "1min",
		},
		{
			name:     "couple of minutes",
			seconds:  120,
			expected: "2min",
		},
		{
			name:     "minutes and seconds",
			seconds:  150,
			expected: "2min 30s",
		},
		{
			name:     "exactly one hour",
			seconds:  3600,
			expected: "1h",
		},
		{
			name:     "hour and minute",
			seconds:  3660,
			expected: "1h 1min",
		},
		{
			name:     "hour and multiple minutes",
			seconds:  3720,
			expected: "1h 2min",
		},
		{
			name:     "more than 24 hours",
			seconds:  93000, // 25h 50min
			expected: "25h 50min",
		},
		{
			name:     "zero seconds",
			seconds:  0,
			expected: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.seconds)
			if result != tt.expected {
				t.Errorf("FormatDuration(%d) = %q, expected %q", tt.seconds, result, tt.expected)
			}
		})
	}
}
