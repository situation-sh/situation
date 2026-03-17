package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func ErrorCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return fmt.Errorf("%s", msg)
	}
}
