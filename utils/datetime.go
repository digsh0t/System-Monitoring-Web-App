package utils

import "time"

// Time format YYYY-MM-DD hh:mm:ss
// Example: 2021-08-25 02:23:35
func GetCurrentDateTime() string {
	currentTime := time.Now()
	datetime := currentTime.Format("2006-01-02 15:04:05")
	return datetime
}
