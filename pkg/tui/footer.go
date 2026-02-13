package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Tab        key.Binding
	Screenshot key.Binding
	Quit       key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Up, k.Down, k.Screenshot, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},                // first column
		{k.Tab, k.Screenshot, k.Quit}, // second column
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next zone"),
	),
	Screenshot: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "take screenshot"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type FooterModel struct {
	SizedModel

	help help.Model
	keys keyMap
}

var helpKeyStyle = lipgloss.NewStyle().Foreground(AccentColor).Faint(true)
var helpDescStyle = lipgloss.NewStyle().Faint(true)
var helpSepStyle = lipgloss.NewStyle().Faint(true)

var helpStyle = help.Styles{
	ShortKey:       helpKeyStyle,
	ShortDesc:      helpDescStyle,
	ShortSeparator: helpSepStyle,
	Ellipsis:       helpSepStyle,
	FullKey:        helpKeyStyle,
	FullDesc:       helpDescStyle,
	FullSeparator:  helpSepStyle,
}

func NewFooterModel() FooterModel {
	h := help.New()
	h.Styles = helpStyle
	return FooterModel{
		help: h,
		keys: keys,
	}
}

func (m FooterModel) SetSize(width, height int) {
	m.help.Width = width
}

func (m FooterModel) Init() tea.Cmd {
	return nil
}

func (m FooterModel) Update(msg tea.Msg) (FooterModel, tea.Cmd) {
	var cmd tea.Cmd
	m.help, cmd = m.help.Update(msg)
	return m, cmd
}

func (m FooterModel) View() string {
	return m.help.View(m.keys)
}
