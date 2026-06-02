package common

import (
	"fmt"
	"regexp"
	"strings"
)

var hexColorPattern = regexp.MustCompile(`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

// NormalizeHexColor trims a color and ensures it is a Blue API-compatible hex string.
func NormalizeHexColor(color string) (string, error) {
	color = strings.TrimSpace(color)
	if color == "" {
		return "", fmt.Errorf("color is required")
	}

	if !hexColorPattern.MatchString(color) {
		return "", fmt.Errorf("invalid color %q: expected #rgb, #rrggbb, or #rrggbbaa", color)
	}

	if !strings.HasPrefix(color, "#") {
		color = "#" + color
	}

	return color, nil
}
