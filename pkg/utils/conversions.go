package utils

import (
	"log"
	"strconv"
	"strings"
	"time"
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
		result = result * 100
	}

	return int64(result)
}

func TimeToUtc(t *time.Time) time.Time {
	utc, _ := time.LoadLocation("UTC")
	newTime := t.In(utc)
	return newTime
}

func TimeToLocal(t *time.Time) time.Time {
	loc := time.Local
	newTime := t.In(loc)
	return newTime
}
