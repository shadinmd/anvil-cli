package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var Color = struct {
	Primary    color.Color
	Secondary  color.Color
	Success    color.Color
	Error      color.Color
	Foreground color.Color
}{
	Primary:    lipgloss.Color("#7aa2f7"), // Tokyo Night Blue
	Secondary:  lipgloss.Color("#565f89"), // Tokyo Night Grey/Comments
	Success:    lipgloss.Color("#9ece6a"), // Tokyo Night Green
	Error:      lipgloss.Color("#f7768e"), // Tokyo Night Red
	Foreground: lipgloss.Color("#c0caf5"), // Tokyo Night Foreground
}
