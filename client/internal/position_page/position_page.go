package positionpage

import (
	"net/http"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

type PositionPage struct {
	BaseModel basemodel.BaseModel
	Position  *messages.Position
}

func NewPositionPage(client *http.Client) PositionPage {
	return PositionPage{
		BaseModel: basemodel.BaseModel{Client: client},
	}
}

func (p PositionPage) Init() tea.Cmd {
	return nil
}

func (p PositionPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return p, tea.Quit

		case "esc":
			return p, func() tea.Msg {
				return messages.SmartPageSwitchMsg{
					Page: messages.ProfilePageNumber,
				}
			}
		}
	}

	return p, nil
}

func (p PositionPage) View() string {
	return ""
}
