package loginpage

import (
	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoginPage struct {
	BaseModel basemodel.BaseModel
	email     textinput.Model
	password  textinput.Model
	help      help.Model
	cursor    int
	typing    bool
}

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Unfucus key.Binding
	Submit  key.Binding
	Help    key.Binding
	Quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Submit},
		{k.Unfucus, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "ctrl+k"),
		key.WithHelp("↑/ctrl+k", "move up (only when typing)"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "ctrl+j"),
		key.WithHelp("↓/ctrl+j", "move down (only when typing)"),
	),
	Unfucus: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "stop typing"),
	),
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "continue typing / submit"),
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

func NewLoginPage() LoginPage {
	email := textinput.New()
	email.Placeholder = "email"
	email.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	email.Focus()

	password := textinput.New()
	password.Placeholder = "password"
	password.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	password.Blur()

	help := help.New()

	return LoginPage{
		email:    email,
		password: password,
		help:     help,
		cursor:   0,
		typing:   true,
	}
}

func (l LoginPage) Init() tea.Cmd {
	return textinput.Blink
}

func (l LoginPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if l.typing {
			switch {
			case key.Matches(msg, keys.Down):
				if l.cursor < 1 {
					l.cursor++
				}
			case key.Matches(msg, keys.Up):
				if l.cursor > 0 {
					l.cursor--
				}
			case key.Matches(msg, keys.Submit):
				return l, l.submit
			case key.Matches(msg, keys.Unfucus):
				l.typing = !l.typing
			}
		} else {
			switch {
			case key.Matches(msg, keys.Help):
				l.help.ShowAll = !l.help.ShowAll
			case key.Matches(msg, keys.Submit):
				l.typing = !l.typing
			case key.Matches(msg, keys.Quit):
				return l, tea.Quit
			}
		}
	}

	if l.typing {
		if l.cursor == 0 {
			l.email.Focus()
			l.password.Blur()
		} else {
			l.email.Blur()
			l.password.Focus()
		}
	} else {
		l.email.Blur()
		l.password.Blur()
	}

	var cmd tea.Cmd
	if l.cursor == 0 {
		l.email, cmd = l.email.Update(msg)
	} else {
		l.password, cmd = l.password.Update(msg)
	}

	return l, cmd
}

func (l LoginPage) View() string {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(32)

	focusedStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#828282")).
		Padding(0, 1).
		Width(32)

	emailView, passwordView := "", ""
	if l.typing {
		if l.email.Focused() {
			emailView = focusedStyle.Render(l.email.View())
			passwordView = inputStyle.Render(l.password.View())
		} else {
			emailView = inputStyle.Render(l.email.View())
			passwordView = focusedStyle.Render(l.password.View())
		}
	} else {
		emailView = inputStyle.Render(l.email.View())
		passwordView = inputStyle.Render(l.password.View())
	}

	ui := lipgloss.JoinVertical(
		lipgloss.Center,
		emailView,
		passwordView,
	)

	return lipgloss.Place(
		l.BaseModel.Width,
		l.BaseModel.Height-3,
		lipgloss.Center,
		lipgloss.Center,
		ui,
	) + "\n" + l.help.View(keys)
}

func (l LoginPage) submit() tea.Msg {
	return messages.PageSwitchMsg{
		Page: messages.WatchlistPageNumber,
	}
}
