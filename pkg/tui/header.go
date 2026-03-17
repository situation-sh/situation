package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var headerStyle = lipgloss.NewStyle().
	Faint(true)
	// Background(PrimaryMutedColor).PaddingLeft(1)

type HeaderModel struct {
	SizedModel
}

func (m HeaderModel) Init() tea.Cmd {
	// return tea.RequestWindowSize
	return nil
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
