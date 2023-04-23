package stringutil

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FormatNumber formats a number to a string with a comma separator
func FormatNumber(number int) string {
	numberStr := strconv.Itoa(number)
	result := ""

	for i, j := len(numberStr)-1, 0; i >= 0; i, j = i-1, j+1 {
		if j%3 == 0 && j != 0 {
			result = "," + result
		}
		result = string(numberStr[i]) + result
	}

	return result
}

// FormatMetricNumber formats a number to a string with a metric suffix
func FormatMetricNumber(number int) string {
	if number < 1000 {
		return strconv.Itoa(number)
	}
	suffixes := []string{"k", "m", "b", "t"}
	num := float64(number)
	var suffix string
	for i := len(suffixes) - 1; i >= 0; i-- {
		divisor := math.Pow(10, float64((i+1)*3))
		if num >= divisor {
			suffix = suffixes[i]
			num /= divisor
			break
		}
	}
	return fmt.Sprintf("%.0f%s", num, suffix)
}

// ParseMetricNumber parses a metric number string to an integer
func ParseMetricNumber(metricNumber string) (int, error) {
	suffixes := []string{"k", "m", "b", "t"}
	var multiplier float64 = 1
	for i, suffix := range suffixes {
		if strings.HasSuffix(metricNumber, suffix) {
			multiplier = math.Pow(10, float64((i+1)*3))
			metricNumber = strings.TrimSuffix(metricNumber, suffix)
			break
		}
	}
	number, err := strconv.ParseFloat(metricNumber, 64)
	if err != nil {
		return 0, err
	}
	return int(number * multiplier), nil
}