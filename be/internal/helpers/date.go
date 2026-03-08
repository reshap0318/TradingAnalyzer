package helpers

import (
	"time"
)

const (
	DateTimeLayout = "2006-01-02 15:04:05"
	DateLayout     = "2006-01-02"
	TimeLayout     = "15:04:05"
)

// FormatDateTime formats time.Time to datetime string (YYYY-MM-DD HH:MM:SS)
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeLayout)
}

// FormatDateTimePtr formats *time.Time to string pointer (returns nil if input is nil)
func FormatDateTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format(DateTimeLayout)
	return &formatted
}

// FormatDate formats time.Time to date string (YYYY-MM-DD)
func FormatDate(t time.Time) string {
	return t.Format(DateLayout)
}

// FormatTime formats time.Time to time string (HH:MM:SS)
func FormatTime(t time.Time) string {
	return t.Format(TimeLayout)
}

// ParseDateTime parses datetime string to time.Time
func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(DateTimeLayout, s)
}

// ParseDate parses date string to time.Time
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateLayout, s)
}

// Now returns current time
func Now() time.Time {
	return time.Now()
}

// GetWIBLocation returns the WIB (Western Indonesian Time) location
func GetWIBLocation() *time.Location {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return loc
}

// ToWIB converts time to WIB (Western Indonesian Time)
func ToWIB(t time.Time) time.Time {
	return t.In(GetWIBLocation())
}

// FormatWIB formats time to WIB string with custom format
func FormatWIB(t time.Time, format string) string {
	return t.In(GetWIBLocation()).Format(format)
}

// FormatWIBDefault formats time to WIB with default format: "02/01/2006, 15.04.05"
func FormatWIBDefault(t time.Time) string {
	return FormatWIB(t, "02/01/2006, 15.04.05")
}

// GetCurrentWIB returns current time in WIB
func GetCurrentWIB() time.Time {
	return time.Now().In(GetWIBLocation())
}

// GetCurrentDateTime returns current datetime as string in "2006-01-02 15:04:05" format
func GetCurrentDateTime() string {
	return time.Now().Format(DateTimeLayout)
}
