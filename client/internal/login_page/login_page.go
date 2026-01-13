package loginpage

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	"github.com/Phantomvv1/KayTrade/client/internal/requests"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoginPage struct {
	BaseModel       basemodel.BaseModel
	email           textinput.Model
	password        textinput.Model
	help            help.Model
	cursor          int
	typing          bool
	viewingPassword bool
}

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Unfucus key.Binding
	Submit  key.Binding
	Switch  key.Binding
	View    key.Binding
	SignUp  key.Binding
	Help    key.Binding
	Quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Unfucus, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit, k.Submit},
		{k.Up, k.Down, k.Unfucus},
		{k.View, k.Switch, k.SignUp},
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
	Switch: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch input field"),
	),
	View: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "view password"),
	),
	SignUp: key.NewBinding(
		key.WithKeys("s", "S"),
		key.WithHelp("s", "sign up"),
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

func NewLoginPage(client *http.Client, tokenStore *basemodel.TokenStore) LoginPage {
	email := textinput.New()
	email.Placeholder = "email"
	email.Width = 25
	email.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	email.Focus()

	password := textinput.New()
	password.Placeholder = "password"
	password.Width = 25
	password.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	password.Blur()

	help := help.New()

	return LoginPage{
		email:     email,
		password:  password,
		help:      help,
		cursor:    0,
		typing:    true,
		BaseModel: basemodel.BaseModel{Client: client, TokenStore: tokenStore},
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
			case key.Matches(msg, keys.View):
				if l.password.Focused() && !l.viewingPassword {
					l.password.EchoMode = textinput.EchoNormal
					l.viewingPassword = !l.viewingPassword
				} else if l.password.Focused() && l.viewingPassword {
					l.password.EchoMode = textinput.EchoPassword
					l.viewingPassword = !l.viewingPassword
				}
			case key.Matches(msg, keys.Switch):
				if l.cursor == 0 {
					l.cursor++
				} else {
					l.cursor--
				}
			}
		} else {
			switch {
			case key.Matches(msg, keys.Help):
				l.help.ShowAll = !l.help.ShowAll
			case key.Matches(msg, keys.Submit):
				l.typing = !l.typing
			case key.Matches(msg, keys.SignUp):
				l.typing = true
				return l, func() tea.Msg {
					return messages.PageSwitchMsg{
						Page: messages.SignUpPageNumber,
					}
				}
			case key.Matches(msg, keys.Quit):
				return l, func() tea.Msg {
					return messages.QuitMsg{}
				}
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
	info := map[string]string{
		"email":    l.email.Value(),
		"password": l.password.Value(),
	}

	reqBody, err := json.Marshal(info)
	if err != nil {
		log.Println(err)
		return messages.PageSwitchMsg{
			Err:  err,
			Page: messages.ErrorPageNumber,
		}
	}

	reader := bytes.NewReader(reqBody)
	body, err := requests.MakeRequest(http.MethodPost, requests.BaseURL+"/log-in", reader, l.BaseModel.Client, &basemodel.TokenStore{Token: ""})
	if err != nil {
		log.Println(err)
		return messages.PageSwitchMsg{
			Err:  err,
			Page: messages.ErrorPageNumber,
		}
	}

	info = nil
	err = json.Unmarshal(body, &info)
	if err != nil {
		log.Println(err)
		return messages.PageSwitchMsg{
			Err:  err,
			Page: messages.ErrorPageNumber,
		}
	}

	l.password.SetValue("")

	return messages.LoginSuccessMsg{
		Token: info["token"],
		Page:  messages.WatchlistPageNumber,
	}
}

func (l *LoginPage) Reload() {
	l.email.SetValue("")
	l.password.SetValue("")
	l.cursor = 0
	l.typing = true
}
