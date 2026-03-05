package utils

import "time"

func RecentYears(count int) []int {
	if count <= 0 {
		return nil
	}

	currentYear := time.Now().Year()
	years := make([]int, 0, count)
	for i := 0; i < count; i++ {
		years = append(years, currentYear-i)
	}

	return years
}
