// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseTimeParam parses time parameters supporting multiple formats:
// - "now" for current time
// - Unix timestamps (e.g., "1704067200")
// - Relative time with flexible formats:
//   - Short: "1h", "30m", "7d", "5s", "1w"
//   - Long: "5min", "5mins", "5minute", "5minutes"
//   - Long: "2hr", "2hrs", "2hour", "2hours"
//   - Long: "3day", "3days"
//   - Long: "1week", "1weeks"
//   - With spaces: "5 minutes", "2 hours"
//   - With minus prefix: "-5m", "-2h" (treated same as "5m", "2h")
//
// - ISO date strings (e.g., "2024-01-01T00:00:00Z")
func ParseTimeParam(timeStr string) (time.Time, error) {
	// Handle "now" (case-insensitive)
	if strings.ToLower(timeStr) == "now" {
		return time.Now(), nil
	}

	// Try parsing as Unix timestamp
	if matched, _ := regexp.MatchString(`^\d+$`, timeStr); matched {
		timestamp, err := strconv.ParseInt(timeStr, 10, 64)
		if err == nil {
			return time.Unix(timestamp, 0), nil
		}
	}

	// Try parsing relative time with flexible formats
	// Supports: 5m, 5min, 5mins, 5minute, 5minutes, 5 minutes, -5m, etc.
	// Case-insensitive to handle MIN, Hour, HOURS, etc.
	re := regexp.MustCompile(`(?i)^-?(\d+)\s*(s|sec|secs|second|seconds|m|min|mins|minute|minutes|h|hr|hrs|hour|hours|d|day|days|w|week|weeks)$`)
	matches := re.FindStringSubmatch(timeStr)
	if len(matches) == 3 {
		value, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time value: %w", err)
		}

		unit := strings.ToLower(matches[2])
		var duration time.Duration

		// Map all variations to their base duration
		switch unit {
		case "s", "sec", "secs", "second", "seconds":
			duration = time.Duration(value) * time.Second
		case "m", "min", "mins", "minute", "minutes":
			duration = time.Duration(value) * time.Minute
		case "h", "hr", "hrs", "hour", "hours":
			duration = time.Duration(value) * time.Hour
		case "d", "day", "days":
			duration = time.Duration(value) * 24 * time.Hour
		case "w", "week", "weeks":
			duration = time.Duration(value) * 7 * 24 * time.Hour
		}

		return time.Now().Add(-duration), nil
	}

	// Try parsing as ISO date
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
}

// ParseTimeToUnix parses time string and returns Unix timestamp in seconds
func ParseTimeToUnix(timeStr string) (int64, error) {
	t, err := ParseTimeParam(timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// ParseTimeToUnixMilli parses time string and returns Unix timestamp in milliseconds.
// Uses Unix()*1000 instead of UnixMilli() to produce second-aligned timestamps.
// time.Now() and duration arithmetic produce nanosecond precision, and UnixMilli()
// preserves the sub-second component which some Datadog APIs reject or misinterpret.
func ParseTimeToUnixMilli(timeStr string) (int64, error) {
	t, err := ParseTimeParam(timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix() * 1000, nil
}
