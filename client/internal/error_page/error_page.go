package errorpage

import (
	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ErrorPage struct {
	baseModel basemodel.BaseModel
	Err       error
}

func (e ErrorPage) Init() tea.Cmd {
	return nil
}

func (e ErrorPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.baseModel.Width, e.baseModel.Height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return e, tea.Quit
		default:
			return e, tea.Quit
		}
	}

	return e, nil
}

func (e ErrorPage) View() string {
	title := lipgloss.PlaceHorizontal(e.baseModel.Width, lipgloss.Center, e.Err.Error())
	subtitle := lipgloss.PlaceHorizontal(e.baseModel.Width, lipgloss.Center, "Press any key to continue • Press q to quit")
	ui := lipgloss.JoinVertical(lipgloss.Center, title, subtitle)

	return lipgloss.Place(e.baseModel.Width, e.baseModel.Height, lipgloss.Center, lipgloss.Center, ui)
}
