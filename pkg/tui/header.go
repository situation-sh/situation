package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var headerStyle = lipgloss.NewStyle().
	Foreground(InvPrimaryColor)
	// Background(PrimaryMutedColor).PaddingLeft(1)

type HeaderModel struct {
	SizedModel
}

func (m HeaderModel) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m HeaderModel) Update(msg tea.Msg) (HeaderModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

func (m HeaderModel) View() string {
	return headerStyle.Width(m.width).Height(1).Render("Situation - Explore")
}
