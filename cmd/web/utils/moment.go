package utils

import "time"

func MinutesFromNow(targetStr string) int {
	targetTime, err := time.Parse(time.RFC3339, targetStr)
	if err != nil {
		panic(err)
	}

	now := time.Now().In(targetTime.Location())
	diff := targetTime.Sub(now)
	minutes := int(diff.Minutes())
	if minutes < 0 {
		minutes = -minutes
	}

	return minutes
}
