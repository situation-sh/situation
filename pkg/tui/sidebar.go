package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/situation-sh/situation/pkg/models"
)

var sidebarStyle = lipgloss.NewStyle().
	Foreground(PrimaryColor).
	Padding(0).
	BorderStyle(TitleBorder(lipgloss.NormalBorder(), "Zones"))

var itemStyle = lipgloss.NewStyle().
	AlignHorizontal(lipgloss.Center)

var selectedItemStyle = lipgloss.NewStyle().
	Bold(true).
	Background(AccentColor).
	AlignHorizontal(lipgloss.Center)

type newSubnetMsg struct {
	subnet *models.Subnetwork
}

type SidebarModel struct {
	SizedModel

	selected int
	subnets  []*models.Subnetwork
}

func NewSidebarModel() *SidebarModel {
	return &SidebarModel{}
}

func (m *SidebarModel) SetSize(width, height int) {
	// remove border
	m.width = width - 2
	m.height = height - 2
}

func (m *SidebarModel) SetSubnets(subnets []*models.Subnetwork) *models.Subnetwork {
	m.subnets = subnets
	if len(subnets) > 0 {
		return subnets[0]
	}
	return nil
}

func (m *SidebarModel) Next() tea.Msg {
	if len(m.subnets) == 0 {
		return nil
	}
	m.selected = (m.selected + 1) % len(m.subnets)
	return newSubnetMsg{subnet: m.subnets[m.selected]}

}

func (m *SidebarModel) Init() tea.Cmd {
	return nil
}

func (m *SidebarModel) Update(msg tea.Msg) (*SidebarModel, tea.Cmd) {
	return m, nil
}

func (m *SidebarModel) View() string {
	subnets := make([]string, len(m.subnets))
	for i, s := range m.subnets {
		if i == m.selected {
			subnets[i] = selectedItemStyle.Width(m.width).Render(s.NetworkCIDR)
		} else {
			subnets[i] = itemStyle.Width(m.width).Render(s.NetworkCIDR)
		}
	}
	// subnets = append(subnets, m.help.View(m.keys))
	// help := m.help.View(m.keys)
	content := lipgloss.JoinVertical(lipgloss.Left, subnets...)
	return sidebarStyle.Width(m.width).Height(m.height).Render(content)
}
