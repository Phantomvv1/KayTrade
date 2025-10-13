package landingpage

import (
	"errors"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	figure "github.com/common-nighthawk/go-figure"
)

type LandingPage struct {
	baseModel basemodel.BaseModel
	quitting  bool
}

func (l LandingPage) Init() tea.Cmd {
	return nil
}

func (l LandingPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.baseModel.Width, l.baseModel.Height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			l.quitting = true
			return l, tea.Quit
		default:
			return l, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.ErrorPageNumber,
					Err:  errors.New("Incorrectly passign pages"),
				}
			}
		}
	}
	return l, nil
}

func (l LandingPage) View() string {
	if l.quitting {
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
		l.baseModel.Width, l.baseModel.Height,
		lipgloss.Center, lipgloss.Center,
		content,
	)

	return ui
}
