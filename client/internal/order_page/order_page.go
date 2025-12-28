package orderpage

import (
	"net/http"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

type OrderPage struct {
	BaseModel basemodel.BaseModel
	Order     *messages.Order
}

func NewOrderPage(client *http.Client) OrderPage {
	return OrderPage{
		BaseModel: basemodel.BaseModel{Client: client},
	}
}

func (o OrderPage) Init() tea.Cmd {
	return nil
}

func (o OrderPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return o, tea.Quit

		case "esc":
			return o, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.ProfilePageNumber,
				}
			}
		}
	}
	return o, nil
}

func (o OrderPage) View() string {
	return ""
}
