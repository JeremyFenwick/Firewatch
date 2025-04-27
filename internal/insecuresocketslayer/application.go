package insecuresocketslayer

import (
	"bytes"
	"errors"
)

func MostCommonToy(line []byte) ([]byte, error) {
	// Handle empty input
	if len(bytes.TrimSpace(line)) == 0 {
		return nil, errors.New("empty request line")
	}

	var maxCount int
	var bestToy []byte
	foundValidToy := false

	entries := bytes.Split(line, []byte{','})
	for _, entry := range entries {
		// Skip empty entries
		entry = bytes.TrimSpace(entry)
		if len(entry) == 0 {
			continue
		}

		// Remove trailing newline if present
		if entry[len(entry)-1] == '\n' {
			entry = entry[:len(entry)-1]
		}

		// Find the 'x' separator
		xIndex := bytes.IndexByte(entry, 'x')
		if xIndex <= 0 || xIndex == len(entry)-1 {
			continue // Skip invalid format
		}

		// Parse count
		count, err := parseCount(entry[:xIndex])
		if err != nil {
			continue // Skip invalid count
		}

		// Update maximum if needed
		if !foundValidToy || count > maxCount {
			maxCount = count
			bestToy = entry
			foundValidToy = true
		}
	}

	if !foundValidToy {
		return nil, errors.New("no valid toy entries found")
	}

	return bestToy, nil
}

func parseCount(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, errors.New("invalid toy entry format")
	}

	count := 0
	// Simple overflow prevention
	const maxSafeInt = (1<<31 - 1) / 10

	for _, b := range data {
		if b < '0' || b > '9' {
			return 0, errors.New("invalid toy entry format")
		}

		// Check for potential overflow
		if count > maxSafeInt {
			return 0, errors.New("count too large")
		}

		count = count*10 + int(b-'0')
	}

	return count, nil
}
