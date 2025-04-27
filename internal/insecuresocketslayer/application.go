package insecuresocketslayer

import (
	"fmt"
	"strconv"
	"strings"
)

// Returns the most common toy and the data up to the newline read
func MostCommonToy(toys string) (string, error) {
	// Split the toys by comma
	toyList := strings.Split(toys, ",")
	toyMax := 0
	maxToy := ""
	// Find the toy with the highest count
	for _, toy := range toyList {
		toy = strings.TrimSpace(toy)
		toyCount, err := GetToyCount(toy)
		if err != nil {
			return "", fmt.Errorf("invalid toy description: %s", toy)
		}
		if toyCount > toyMax {
			toyMax = toyCount
			maxToy = toy
		}
	}
	// We may need to trim the toy
	maxToy = strings.TrimRight(maxToy, "\n")
	return maxToy, nil
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
