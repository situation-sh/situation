package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func ErrorCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return fmt.Errorf("%s", msg)
	}
}
