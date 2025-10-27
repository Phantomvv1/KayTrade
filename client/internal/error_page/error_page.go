package errorpage

import (
	"log"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ErrorPage struct {
	BaseModel basemodel.BaseModel
	Err       error
	PrevPage  int
}

func (e ErrorPage) Init() tea.Cmd {
	return nil
}

func (e ErrorPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return e, tea.Quit
		default:
			return e, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: e.PrevPage,
				}
			}
		}
	}

	return e, nil
}

func (e ErrorPage) View() string {
	log.Printf("Width: %d, Height: %d . Error page!", e.BaseModel.Width, e.BaseModel.Height)
	content := lipgloss.JoinVertical(lipgloss.Center, e.Err.Error(), "Press any key to continue • Press q to quit")
	ui := lipgloss.JoinVertical(lipgloss.Center, content)

	return lipgloss.Place(e.BaseModel.Width, e.BaseModel.Height, lipgloss.Center, lipgloss.Center, ui)
}
