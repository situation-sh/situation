package tui

import (
	"fmt"
	"path"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/situation-sh/situation/pkg/models"
)

var cardBaseStyle = lipgloss.
	NewStyle().
	Padding(0, 1).
	BorderStyle(TitleBorder(lipgloss.NormalBorder(), "Node"))

var keyStyle = lipgloss.NewStyle().Faint(true).Align(lipgloss.Right).PaddingRight(1)

func tableStyle(row, col int) lipgloss.Style {
	switch {
	case col == 0:
		return keyStyle.Width(14)
	default:
		return lipgloss.NewStyle().PaddingRight(1)
	}
}

func baseTable() *table.Table {
	return table.New().
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false).
		StyleFunc(tableStyle)
}

type CardModel struct {
	SizedModel

	nic       *models.NetworkInterface
	addr      string
	endpoints []*models.ApplicationEndpoint
}

func NewCardModel() *CardModel {
	return &CardModel{}
}

func (m *CardModel) ports() [][]string {
	out := make([][]string, 0)
	for _, e := range m.endpoints {
		if e.Addr == m.addr {
			proto := ""
			if len(e.ApplicationProtocols) > 0 {
				proto = e.ApplicationProtocols[0]
			}
			out = append(out,
				[]string{
					fmt.Sprintf("%s/%d", e.Protocol, e.Port),
					proto,
					fmt.Sprintf("(%d)", len(e.IncomingFlows)),
				},
			)

		}
	}
	return out
}

func (m *CardModel) flows() [][]string {
	tmp := make(map[string][][]string)
	if m.nic == nil || m.nic.OutgoingFlows == nil {
		return nil
	}
	for _, flow := range m.nic.OutgoingFlows {
		if flow.SrcAddr == m.addr {
			saas := ""
			if flow.DstEndpoint.SaaS != "" {
				saas = fmt.Sprintf("(%s)", flow.DstEndpoint.SaaS)
			}
			tmp[flow.SrcApplication.Name] = append(
				tmp[flow.SrcApplication.Name],
				[]string{
					path.Base(flow.SrcApplication.Name),
					"→",
					fmt.Sprintf("%s:%d", flow.DstEndpoint.Addr, flow.DstEndpoint.Port),
					saas,
				},
			)
		}
	}

	out := make([][]string, 0)
	for _, endpoints := range tmp {
		if len(endpoints) > 3 {
			e := endpoints[0]
			e[2] = fmt.Sprintf("(%d)", len(endpoints))
			e[3] = ""
			out = append(out, e)
		} else {
			out = append(out, endpoints...)
		}

	}
	slices.SortFunc(out, func(a, b []string) int {
		if a[0] < b[0] {
			return -1
		} else if a[0] > b[0] {
			return 1
		}
		return 0
	})

	return out
}

func (m *CardModel) systemWidth() int {
	ratio := 1. / 2.
	if m.width > 100 {
		ratio = 1. / 3.
	}
	w := int(ratio * float64(m.width-2))
	if w > 34 {
		w = 34
	}
	return w
}

func (m *CardModel) SetSource(nic *models.NetworkInterface, addr string) {
	m.nic = nic
	m.addr = addr
	m.endpoints = make([]*models.ApplicationEndpoint, 0)
	if nic != nil {
		for _, e := range nic.Endpoints {
			// accept also empty address to show all endpoints for the NIC
			if addr == "" || e.Addr == addr {
				m.endpoints = append(m.endpoints, e)
			}
		}
	}
}

func (m *CardModel) SetSize(width, height int) {
	// remove border
	m.width = width - 2
	m.height = height - 2
}

func (m *CardModel) Init() tea.Cmd {
	return nil
}

func (m *CardModel) Update(msg tea.Msg) (*CardModel, tea.Cmd) {
	return m, nil
}

func (m *CardModel) View() string {
	systemRows := [][]string{
		{"Hostname", ""},
		{"Platform", ""},
		{"Distribution", ""},
		{"Version", ""},
		{"Chassis", ""},
	}
	if m.nic != nil && m.nic.Machine != nil {
		systemRows[0][1] = m.nic.Machine.Hostname
		systemRows[1][1] = m.nic.Machine.Platform
		systemRows[2][1] = m.nic.Machine.Distribution
		systemRows[3][1] = m.nic.Machine.DistributionVersion
		systemRows[4][1] = m.nic.Machine.Chassis
	}

	bold := lipgloss.NewStyle().Bold(true)
	h := m.height - 3

	// sys
	t := baseTable().Rows(systemRows...).Width(m.systemWidth()).Height(h)
	system := lipgloss.JoinVertical(lipgloss.Left, bold.Render("System"), t.String())

	// endpoints
	e := baseTable().StyleFunc(func(row, col int) lipgloss.Style {
		if col == 1 {
			return lipgloss.NewStyle().Faint(true).Padding(0, 1)
		}
		return lipgloss.NewStyle()
	}).Rows(m.ports()...).Height(h)
	endpoints := lipgloss.JoinVertical(
		lipgloss.Left,
		bold.Width(28).Render("Endpoints (incoming flows)"),
		e.String(),
	)

	// flows

	flows := ""
	if m.width > 100 {
		flows = lipgloss.JoinVertical(
			lipgloss.Left,
			bold.Render("Flows (outgoing)"),
			baseTable().StyleFunc(func(row, col int) lipgloss.Style {
				if col == 1 {
					return lipgloss.NewStyle().Padding(0, 1)
				}
				if col == 3 {
					return lipgloss.NewStyle().Faint(true).Padding(0, 0, 0, 1)
				}
				return lipgloss.NewStyle()
			}).Rows(m.flows()...).Height(h).String(),
		)
	}

	return cardBaseStyle.
		Width(m.width).
		Height(m.height).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, system, endpoints, flows))
}
