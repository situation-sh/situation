package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func TitleBorder(border lipgloss.Border, title string) lipgloss.Border {
	// put a high number to ensure the title is writtent only once
	border.Top = fmt.Sprintf("%s %s %s", border.Top, title, strings.Repeat(border.Top, 256))
	return border
}
