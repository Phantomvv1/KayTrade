package landingpage

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
)

type Model struct {
	width, height int
	quitting      bool
}

func NewLandingPageModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			// "continue" → quit landing page for now
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	title := figure.NewFigure("KayTrade", "", true).String()

	// Styles
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")). // gray
		Bold(true)

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")). // gray
		Italic(true).
		Render("Press any key to continue • Press q to quit")

	content := lipgloss.JoinVertical(lipgloss.Center, titleStyle.Render(title), subtitle)

	ui := lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)

	return ui
}
