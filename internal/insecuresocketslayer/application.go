package insecuresocketslayer

import (
	"fmt"
	"strconv"
	"strings"
)

// Returns the most common toy and the data up to the newline read
func MostCommonToy(data []byte) (string, int, error) {
	toys := string(data)
	newLineIndex := strings.Index(toys, "\n")
	if newLineIndex == -1 {
		return "", 0, fmt.Errorf("no newline found in data")
	}
	toys = toys[:newLineIndex]
	// Split the toys by comma
	toyList := strings.Split(toys, ",")
	toyMax := 0
	maxToy := ""
	// Find the toy with the highest count
	for _, toy := range toyList {
		toy = strings.TrimSpace(toy)
		toyCount, err := GetToyCount(toy)
		if err != nil {
			return "", 0, fmt.Errorf("invalid toy description: %s", toy)
		}
		if toyCount > toyMax {
			toyMax = toyCount
			maxToy = toy
		}
	}
	return maxToy, newLineIndex + 1, nil
}

func GetToyCount(toyDesc string) (int, error) {
	xIndex := strings.Index(toyDesc, "x")
	if xIndex == -1 {
		return 0, fmt.Errorf("invalid toy description: %s", toyDesc)
	}
	toyCount, err := strconv.Atoi(toyDesc[:xIndex])
	if err != nil {
		return 0, fmt.Errorf("invalid toy count: %s", toyDesc[:xIndex])
	}
	return toyCount, nil
}
