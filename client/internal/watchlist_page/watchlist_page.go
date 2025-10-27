package watchlistpage

import (
	"net/http"
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WatchlistPage struct {
	BaseModel basemodel.BaseModel
	titleBar  string
	companies []string
	cursor    int
	help      help.Model
}

func NewWatchlistPage(client *http.Client) WatchlistPage {
	return WatchlistPage{
		BaseModel: basemodel.BaseModel{Client: client},
		help:      help.New(),
		cursor:    0,
		titleBar:  "String",
	}
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Help   key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select the company"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (w WatchlistPage) Init() tea.Cmd {
	return nil
}

func (w WatchlistPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return w, tea.Quit
		case key.Matches(msg, keys.Up):
			if w.cursor >= 0 {
				w.cursor--
			}
		case key.Matches(msg, keys.Down):
			if w.cursor < len(w.companies)-1 {
				w.cursor++
			}
		case key.Matches(msg, keys.Select):
			return w, tea.Quit
		case key.Matches(msg, keys.Help):
			w.help.ShowAll = !w.help.ShowAll
		}
	}

	return w, nil
}
func (w WatchlistPage) View() string {
	header := w.titleBar
	helpView := w.help.View(keys)
	height := 8 - strings.Count(helpView, "\n")

	header = lipgloss.PlaceHorizontal(w.BaseModel.Width, lipgloss.Center, header)
	info := lipgloss.PlaceHorizontal(w.BaseModel.Width, lipgloss.Center, "Some company info")

	return header + info + strings.Repeat("\n", height) + helpView
}
