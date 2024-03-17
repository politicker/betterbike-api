package domain

import (
	"fmt"
	"time"
)

func timeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)

	switch {
	case seconds < 60:
		return fmt.Sprintf("%d seconds ago", seconds)
	case minutes < 60:
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours < 24:
		return fmt.Sprintf("%d hours ago", hours)
	default:
		return fmt.Sprintf("%d days ago", days)
	}
}
