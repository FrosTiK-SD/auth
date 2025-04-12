package util

import (
	"strconv"
	"strings"
)

func getRollNumberBounds(searchDigits string) (int, int) {
	paddingLength := 8 - len(searchDigits)

	if paddingLength <= 0 {
		num, _ := strconv.Atoi(searchDigits[:8])
		return num, num
	}

	// Create lower bound by padding with zeros
	lowerStr := searchDigits + strings.Repeat("0", paddingLength)

	// Create upper bound by padding with nines
	upperStr := searchDigits + strings.Repeat("9", paddingLength)

	// Convert to integers
	lowerBound, _ := strconv.Atoi(lowerStr)
	upperBound, _ := strconv.Atoi(upperStr)

	return lowerBound, upperBound
}
