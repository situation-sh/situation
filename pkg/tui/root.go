package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
)

// okMsg signals that async data loading has completed
// and the view needs to be re-rendered.
type okMsg struct{}

type RootModel struct {
	ctx     context.Context
	storage *store.BunStorage
	header  HeaderModel
	sidebar *SidebarModel
	table   *TableModel
	card    *CardModel
	footer  FooterModel
}

func NewRootModel(ctx context.Context, storage *store.BunStorage) RootModel {
	return RootModel{
		ctx:     ctx,
		storage: storage,
		table:   NewTableModel(),
		header:  HeaderModel{},
		sidebar: NewSidebarModel(),
		card:    NewCardModel(),
		footer:  NewFooterModel(),
	}
}

func (m RootModel) FetchSubnets() error {
	subnets := make([]*models.Subnetwork, 0)
	err := m.storage.DB().
		NewSelect().
		Model(&subnets).
		Scan(m.ctx)
	if err != nil {
		return err
	}
	subnet := m.sidebar.SetSubnets(subnets)
	if subnet != nil {
		return m.FetchNICs(subnet)
	}
	return nil
}

func (m RootModel) FetchNICs(subnet *models.Subnetwork) error {
	if subnet == nil {
		return nil
	}
	nics := make([]*models.NetworkInterface, 0)
	err := m.storage.DB().NewSelect().
		Model(&nics).
		Relation("Machine").
		Relation("Endpoints").
		Relation("Endpoints.IncomingFlows").
		Relation("OutgoingFlows").
		Relation("OutgoingFlows.SrcApplication").
		Relation("OutgoingFlows.DstEndpoint").
		Where("network_interface.id IN (?)",
			m.storage.DB().NewSelect().
				Model((*models.NetworkInterfaceSubnet)(nil)).
				Column("network_interface_id").
				Where("subnetwork_id = ?", subnet.ID),
		).Scan(m.ctx)
	if err != nil {
		return err
	}
	m.table.SetNics(nics, subnet.NetworkCIDR)
	return nil
}

func (m RootModel) Fetch() tea.Cmd {
	return func() tea.Msg {
		err := m.FetchSubnets()
		if err != nil {
			return tea.Quit()
		}
		return okMsg{}
	}
}

func (m RootModel) Init() tea.Cmd {
	return tea.Batch(m.Fetch(), tea.WindowSize())
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	var cmd1, cmd2, cmd3 tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h2 := (msg.Height - 2) / 2
		m.header.SetSize(msg.Width, 1)
		m.sidebar.SetSize(24, msg.Height-2)
		m.table.SetSize(msg.Width-24, h2)
		m.card.SetSize(msg.Width-24, msg.Height-2-h2)
		m.footer.SetSize(msg.Width, 1)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			return m, m.sidebar.Next
		}
	case newSubnetMsg:
		return m, func() tea.Msg {
			err := m.FetchNICs(msg.subnet)
			if err != nil {
				return tea.Quit
			}
			return okMsg{}
		}
	case newNodeMsg:
		m.card.SetSource(msg.nic, msg.addr)
		// return m, nil
	case okMsg:
		// do nothing - just re-render with new data
	}

	// pass message to sub-models (only table can handle them for now)
	m.table, cmd1 = m.table.Update(msg)
	return m, tea.Batch(append(cmds, cmd1, cmd2, cmd3)...)
}

func (m RootModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.header.View(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.sidebar.View(),
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.table.View(),
				m.card.View(),
			),
		),
		m.footer.View(),
	)
}

func (m RootModel) Run() error {
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}
