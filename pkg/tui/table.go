package tui

import (
	"fmt"
	"net"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/situation-sh/situation/pkg/models"
)

var colRatios = []float64{
	2.0 / 9.0,
	2.0 / 9.0,
	2.0 / 9.0,
	1.0 / 3.0,
	9.,
}

func fixedWidth() int {
	total := 0
	for _, ratio := range colRatios {
		total += 2 // remove padding
		if ratio > 1.0 {
			total += int(ratio)
		}
	}
	return total
}

var columns = []table.Column{
	{Title: "Machine", Width: 14},
	{Title: "IP", Width: 16},
	{Title: "MAC", Width: 18},
	{Title: "Vendor", Width: 20},
	{Title: "Endpoints", Width: 10},
}

type newNodeMsg struct {
	nic  *models.NetworkInterface
	addr string
}

type TableModel struct {
	table table.Model
	nics  []*models.NetworkInterface
}

func NewTableModel() *TableModel {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Padding(0, 1, 0, 1)
	s.Cell = s.Cell.Padding(0, 1, 0, 1) // right-left padding only
	s.Selected = s.Selected.Foreground(AccentColor)
	t.SetStyles(s)

	return &TableModel{
		table: t,
		nics:  make([]*models.NetworkInterface, 0),
	}
}

func filterIP(ips []string, cidr string) string {
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return ""
	}
	for _, ip := range ips {
		if n.Contains(net.ParseIP(ip)) {
			return ip
		}
	}
	return ""
}
func (m *TableModel) SetSize(width, height int) {
	cols := m.table.Columns()
	tableWidth := width - 2 // remove borders
	// remove the border width
	w := tableWidth - fixedWidth()
	k := 0
	for i := range cols {
		if colRatios[i] > 1.0 {
			cols[i].Width = int(colRatios[i])
		} else {
			cols[i].Width = int(float64(w) * colRatios[i])
			k += cols[i].Width
		}
	}
	if k > w {
		cols[0].Width -= (k - w)
	}

	m.table.SetHeight(height - 2)
	m.table.SetWidth(width - 2)
}

func (m *TableModel) SetNics(nics []*models.NetworkInterface, cidr string) {
	m.nics = nics
	rows := make([]table.Row, len(nics))
	for i, nic := range nics {
		ip := filterIP(nic.IP, cidr)
		name := ""
		if nic.Machine != nil {
			name = nic.Machine.Hostname
		}
		endpoints := 0
		if len(nic.Endpoints) > 0 {
			for _, ep := range nic.Endpoints {
				if ep.Addr == ip {
					endpoints++
				}
			}
		}
		rows[i] = table.Row{name, ip, nic.MAC, nic.MACVendor, fmt.Sprintf("%d", endpoints)}
	}

	m.table.SetRows(rows)
}

func (m *TableModel) Init() tea.Cmd {
	return nil
}

func (m *TableModel) Update(msg tea.Msg) (*TableModel, tea.Cmd) {
	var cmd tea.Cmd
	// cmd is always nil here
	m.table, cmd = m.table.Update(msg)
	row := m.table.SelectedRow()
	index := m.table.Cursor()
	// WARNING: hardcoded column index

	cmd = func() tea.Msg {
		if len(row) < 2 {
			return nil
		}
		ip := row[1]
		if index < len(m.nics) {
			return newNodeMsg{nic: m.nics[index], addr: ip}
		}
		return nil
	}

	return m, cmd
}

func (m *TableModel) View() string {
	return lipgloss.
		NewStyle().
		BorderStyle(TitleBorder(lipgloss.NormalBorder(), "Data")).
		Render(m.table.View())
}
