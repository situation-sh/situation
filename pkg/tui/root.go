package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/situation-sh/situation/pkg/models"
	"github.com/situation-sh/situation/pkg/store"
)

var pressEnter = lipgloss.NewStyle().Render("Press Enter to continue")

// okMsg signals that async data loading has completed
// and the view needs to be re-rendered.
type okMsg struct{}

// successMsg
type successMsg string

type RootModel struct {
	ctx     context.Context
	storage *store.BunStorage
	header  HeaderModel
	sidebar *SidebarModel
	table   *TableModel
	card    *CardModel
	footer  FooterModel

	err     error
	success string

	width  int
	height int
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

func (m RootModel) Screenshot() tea.Msg {
	name := path.Join(os.TempDir(), fmt.Sprintf("situation-%d.svg", time.Now().Unix()))
	// #nosec G304
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	svg, err := ansi2svg(m.View().Content)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(f, svg); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	// #nosec G302
	if err := os.Chmod(name, 0o644); err != nil {
		return err
	}
	return successMsg(fmt.Sprintf("Screenshot saved to %s", name))
}

func (m RootModel) Init() tea.Cmd {
	// return tea.Batch(m.Fetch(), tea.WindowSize())
	return tea.Batch(m.Fetch())
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd1 tea.Cmd

	if m.err != nil || m.success != "" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.err = nil
				m.success = ""
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case error:
		m.err = msg
		return m, nil
	case successMsg:
		m.success = string(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

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
		case "ctrl+s":
			return m, m.Screenshot
		case "r":
			return m, ErrorCmd("You have pressed 'r'")
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
	return m, cmd1
}

type Viewable string

func (v Viewable) View() string {
	return string(v)
}

func (m RootModel) View() tea.View {
	compositor := lipgloss.NewCompositor()
	// background layer
	bg := lipgloss.NewLayer(
		lipgloss.JoinVertical(lipgloss.Left,
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
		),
	)

	compositor.AddLayers(bg)

	if m.err != nil {
		compositor.AddLayers(m.ErrModal())
	} else if m.success != "" {
		compositor.AddLayers(m.SuccessModal())
	}

	view := tea.NewView(compositor.Render())
	view.AltScreen = true
	view.Cursor = nil
	view.MouseMode = tea.MouseModeNone
	return view
}

func (m RootModel) ErrModal() *lipgloss.Layer {
	if m.err == nil {
		return nil
	}
	msg := lipgloss.JoinVertical(lipgloss.Center, m.err.Error(), "\n", pressEnter)
	return m.Modal(msg,
		func(s lipgloss.Style) lipgloss.Style {
			return s.Background(ErrorBgColor).Foreground(ErrorFgColor)
		},
	)
}

func (m RootModel) SuccessModal() *lipgloss.Layer {
	if m.success == "" {
		return nil
	}
	msg := lipgloss.JoinVertical(lipgloss.Center, m.success, "\n", pressEnter)
	return m.Modal(msg,
		func(s lipgloss.Style) lipgloss.Style {
			return s.BorderStyle(lipgloss.RoundedBorder()).BorderForeground(AccentColor)
		},
	)
}

func (m RootModel) Modal(msg string, opts ...func(s lipgloss.Style) lipgloss.Style) *lipgloss.Layer {
	style := lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(2 * m.width / 5).
		Height(m.height / 3)
	for _, opt := range opts {
		style = opt(style)
	}
	return lipgloss.NewLayer(style.Render(msg)).X(3 * m.width / 10).Y(m.height / 3)
	// return style.Render(msg)
}

func (m RootModel) Run() error {
	_, err := tea.NewProgram(m).Run()
	return err
}
