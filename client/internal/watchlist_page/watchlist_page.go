package watchlistpage

import (
	"errors"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

type WatchlistPage struct {
	BaseModel basemodel.BaseModel
	titleBar  string
	companies []string
	cursor    int
}

func (w WatchlistPage) Init() tea.Cmd {
	return nil
}

func (w WatchlistPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return w, tea.Quit
		case "k", "up":
			if w.cursor >= 0 {
				w.cursor--
			}
		case "j", "down":
			if w.cursor < len(w.companies)-1 {
				w.cursor++
			}
		case "enter":

		default:
			return w, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.ErrorPageNumber,
					Err:  errors.New("Incorrectly passign pages"),
				}
			}
		}
	}
	return w, nil
}
func (w WatchlistPage) View() string {
	return ""
}
