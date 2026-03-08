package helpers

import (
	"math"
	"strconv"
)

// ParseFloat parses string to float64 with specified bitSize
func ParseFloat(s string, bitSize int) (float64, error) {
	return strconv.ParseFloat(s, bitSize)
}

// ParseInt parses string to int64 with specified bitSize
func ParseInt(s string, bitSize int) (int64, error) {
	return strconv.ParseInt(s, 10, bitSize)
}

// ParseBool parses string to bool
func ParseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

// RoundFloat rounds float64 to specified decimal places
func RoundFloat(val float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(val*multiplier) / multiplier
}
