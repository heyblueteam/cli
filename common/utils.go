package common

import "fmt"

// TruncateString truncates a string to the specified length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// PrintSuccess prints a success message with a green checkmark
func PrintSuccess(message string) {
	fmt.Printf("✓ %s\n", message)
}

// PrintError prints an error message with a red X
func PrintError(message string) {
	fmt.Printf("✗ %s\n", message)
}

// PrintInfo prints an info message with an information symbol
func PrintInfo(message string) {
	fmt.Printf("ℹ %s\n", message)
}

// ProjectCategories - Available project categories
var ProjectCategories = []string{
	"GENERAL", "CRM", "MARKETING", "ENGINEERING", "PRODUCT", "SALES",
	"DESIGN", "FINANCE", "HR", "LEGAL", "OPERATIONS", "SUPPORT",
}

// Common project colors
var ProjectColors = map[string]string{
	"blue":   "#3B82F6",
	"red":    "#EF4444",
	"green":  "#10B981",
	"purple": "#8B5CF6",
	"yellow": "#F59E0B",
	"pink":   "#EC4899",
	"indigo": "#6366F1",
	"gray":   "#6B7280",
}

// Available project icons
var ProjectIcons = []string{
	"mdi-cash",
	"mdi-shield-bug-outline",
	"mdi-account-box-outline",
	"mdi-account-group-outline",
	"mdi-alarm-panel-outline",
	"mdi-animation-play-outline",
	"mdi-application-brackets-outline",
	"mdi-archive-arrow-up-outline",
	"mdi-badge-account-horizontal-outline",
	"mdi-bank-outline",
	"mdi-basket-outline",
	"mdi-book-open-outline",
	"mdi-briefcase-variant-outline",
	"mdi-car-outline",
	"mdi-cake-variant-outline",
	"mdi-calendar-account-outline",
	"mdi-camera-outline",
	"mdi-card-account-mail-outline",
	"mdi-cards-club-outline",
	"mdi-cards-heart-outline",
	"mdi-cellphone-basic",
	"mdi-chart-line",
	"mdi-flag-variant-outline",
	"mdi-chat-outline",
	"mdi-cloud-check-outline",
	"mdi-clipboard-list-outline",
	"mdi-clock-time-eight-outline",
	"mdi-video-outline",
	"mdi-gamepad-round-outline",
	"mdi-earth",
	"mdi-image-frame",
	"mdi-laptop",
	"mdi-microphone-outline",
	"mdi-music-note",
	"mdi-cog-outline",
	"mdi-compass-outline",
	"mdi-home-outline",
	"mdi-airplane-takeoff",
	"mdi-gamepad-variant-outline",
	"mdi-key-outline",
	"mdi-folder",
	"mdi-folder-search-outline",
}