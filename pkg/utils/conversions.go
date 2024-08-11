package utils

import (
	"log"
	"strconv"
	"strings"
)

func GetCentsFromString(amount string) int64 {
	var clean string
	clean = strings.TrimSpace(amount)

	hasCents := strings.Index(clean, ".") >= 0

	clean = strings.ReplaceAll(clean, ".", "")
	clean = strings.ReplaceAll(clean, "$", "")
	clean = strings.ReplaceAll(clean, " ", "")

	result, err := strconv.Atoi(clean)
	if err != nil {
		log.Printf("Failed to convert %s (%s) to cents.\n", amount, clean)
	}

	if !hasCents {
		log.Printf("Has Cents\n")
		result = result * 100
	}

	return int64(result)
}
