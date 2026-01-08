package transferspage

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TransfersPage struct {
	BaseModel          basemodel.BaseModel
	amount             textinput.Model
	FundingInformation *messages.FundingInformation
	direction          []string
	directionIdx       int
	cursor             int
	typing             bool
	err                string
	success            string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB88FF")).
			Width(25).
			Align(lipgloss.Center)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)
)

func NewTransfersPage(client *http.Client, tokenStore *basemodel.TokenStore) TransfersPage {
	amount := textinput.New()
	amount.Focus()
	amount.Placeholder = "Amount"
	amount.Width = 27
	amount.CharLimit = 20

	return TransfersPage{
		BaseModel:    basemodel.BaseModel{Client: client, TokenStore: tokenStore},
		amount:       amount,
		direction:    []string{"INCOMING", "OUTGOING"},
		directionIdx: 0,
		cursor:       0,
		typing:       true,
	}
}

func (t TransfersPage) Init() tea.Cmd {
	return textinput.Blink
}

func (t TransfersPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if t.typing {
			switch msg.String() {
			case "esc":
				t.typing = false
				t.amount.Blur()
				return t, nil

			case "ctrl+j", "down":
				t.cursor++
				if t.cursor > 1 {
					t.cursor = 0
				}
				if t.cursor == 0 {
					t.amount.Focus()
				} else {
					t.amount.Blur()
				}
				return t, nil

			case "ctrl+k", "up":
				t.cursor--
				if t.cursor < 0 {
					t.cursor = 1
				}
				if t.cursor == 0 {
					t.amount.Focus()
				} else {
					t.amount.Blur()
				}
				return t, nil

			case "h", "left":
				if t.cursor == 1 {
					t.directionIdx--
					if t.directionIdx < 0 {
						t.directionIdx = len(t.direction) - 1
					}
					return t, nil
				}

			case "l", "right":
				if t.cursor == 1 {
					t.directionIdx++
					if t.directionIdx >= len(t.direction) {
						t.directionIdx = 0
					}
					return t, nil
				}

			case "tab":
				t.cursor++
				if t.cursor > 1 {
					t.cursor = 0
				}

				return t, nil

			case "enter":
				t.err = ""
				t.success = ""
				if err := t.Submit(); err != nil {
					t.err = err.Error()
				} else {
					t.success = "Transfer submitted successfully!"
				}
				return t, nil
			}

			if t.cursor == 0 {
				t.amount, cmd = t.amount.Update(msg)
			}

			return t, cmd
		} else {
			switch msg.String() {
			case "q", "ctrl+c":
				return t, func() tea.Msg {
					return messages.QuitMsg{}
				}

			case "esc":
				t.success = ""
				t.err = ""
				return t, func() tea.Msg {
					return messages.SmartPageSwitchMsg{
						Page: messages.BankRelationshipPageNumber,
					}
				}

			case "enter":
				t.typing = true
				if t.cursor == 0 {
					t.amount.Focus()
				}

				return t, nil
			}
		}
	}

	return t, cmd
}

func (t TransfersPage) View() string {
	header := titleStyle.Render("💸 Transfer Funds")

	var fields []string

	if t.cursor == 0 {
		if t.typing {
			t.amount.Focus()
		} else {
			t.amount.Blur()
		}
	} else {
		t.amount.Blur()
	}

	fields = append(fields, t.renderField("Transfer type", t.FundingInformation.TransferType, false))

	fields = append(fields, t.renderField("Amount", t.amount.View(), t.amount.Focused()))

	fields = append(fields, t.renderField("Direction", t.renderSlider(t.direction, t.directionIdx, t.cursor == 1 && t.typing), t.cursor == 1 && t.typing))

	content := lipgloss.JoinVertical(lipgloss.Center, fields...)

	if t.err != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", errorStyle.Render("❌ "+t.err))
	}
	if t.success != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", successStyle.Render("✓ "+t.success))
	}

	var help string
	if t.typing {
		help = helpStyle.Render("ctrl+j/ctrl+k/↑/↓: navigate | h/l/←/→: change direction | enter: submit | esc: stop typing")
	} else {
		help = helpStyle.Render("enter: start typing | esc: back | q: quit")
	}

	headerHeight := lipgloss.Height(header)
	contentHeight := lipgloss.Height(content)
	helpHeight := lipgloss.Height(help)

	centeredHeader := lipgloss.Place(t.BaseModel.Width, headerHeight, lipgloss.Center, lipgloss.Top, header)
	centeredContent := lipgloss.Place(t.BaseModel.Width, contentHeight, lipgloss.Center, lipgloss.Top, content)
	centeredHelp := lipgloss.Place(t.BaseModel.Width, helpHeight, lipgloss.Center, lipgloss.Top, help)

	finalView := lipgloss.JoinVertical(
		lipgloss.Center,
		centeredHeader,
		strings.Repeat("\n", 16),
		centeredContent,
		"",
		centeredHelp,
	)

	return finalView
}

func (t TransfersPage) renderField(label, value string, focused bool) string {
	styledLabel := labelStyle.Render(label)

	var fieldStyle lipgloss.Style
	if focused {
		fieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Background(lipgloss.Color("#2a2a4e")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FFFF")).
			Padding(0, 1).
			Width(32).
			Align(lipgloss.Center)
	} else {
		fieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(0, 1).
			Width(32).
			Align(lipgloss.Center)
	}

	styledValue := fieldStyle.Render(value)

	// Join label and field vertically, centered
	return lipgloss.JoinVertical(lipgloss.Center, styledLabel, styledValue)
}

func (t TransfersPage) renderSlider(options []string, selectedIdx int, focused bool) string {
	selected := strings.ToUpper(options[selectedIdx])

	if focused {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Align(lipgloss.Center).
			Render("◀ " + selected + " ▶")
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Align(lipgloss.Center).
		Render("◀ " + selected + " ▶")
}

func (t *TransfersPage) Submit() error {
	data := make(map[string]any)

	if t.FundingInformation == nil {
		return errors.New("Error no funding information provided")
	}

	data["transfer_type"] = t.FundingInformation.TransferType
	data["direction"] = t.direction[t.directionIdx]
	data["timing"] = "immediate"

	if t.FundingInformation.TransferType == "ach" && t.FundingInformation.RelationshipId != "" {
		data["relationship_id"] = t.FundingInformation.RelationshipId
	} else if t.FundingInformation.TransferType == "wire" && t.FundingInformation.BankId != "" {
		data["bank_id"] = t.FundingInformation.BankId
	} else {
		return errors.New("Error no relationship or bank found to do this transfer with")
	}

	amountString := strings.TrimSpace(t.amount.Value())
	if amountString == "" {
		return errors.New("Error amount is required")
	}

	amount, err := strconv.Atoi(amountString)
	if err != nil {
		return errors.New("Error invalid amount: must be a number")
	}

	if amount <= 0 {
		return errors.New("Error amount must be greater than 0")
	}

	data["amount"] = amount

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(jsonData)
	_, err = requests.MakeRequest(http.MethodPost, requests.BaseURL+"/transfers", reader, t.BaseModel.Client, t.BaseModel.TokenStore)
	if err != nil {
		return err
	}

	return nil
}
