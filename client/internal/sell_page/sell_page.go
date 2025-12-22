package sellpage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

type SellPage struct {
	BaseModel       basemodel.BaseModel
	Symbol          string
	quantity        textinput.Model
	MaxQuantity     float64
	side            string
	purchaseType    []string
	purchaseTypeIdx int
	timeInForce     []string
	timeInForceIdx  int
	cursor          int
	totalFields     int
	err             string
	success         string
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

func NewSellPage(client *http.Client) SellPage {
	quantity := textinput.New()
	quantity.Placeholder = "Quantity"
	quantity.Width = 28
	quantity.SetValue("1")
	quantity.CharLimit = 10
	quantity.Focus()

	return SellPage{
		BaseModel:       basemodel.BaseModel{Client: client},
		quantity:        quantity,
		side:            "sell",
		purchaseType:    []string{"market", "limit", "stop", "stop_limit", "trailing_stop"},
		purchaseTypeIdx: 0,
		timeInForce:     []string{"day", "gtc", "opg", "cls", "ioc", "fok"},
		timeInForceIdx:  0,
		totalFields:     3,
		cursor:          0,
	}
}

func (s SellPage) Init() tea.Cmd {
	return textinput.Blink
}

func (s SellPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "j", "down":
			s.err = ""
			s.success = ""
			s.cursor++
			if s.cursor >= s.totalFields {
				s.cursor = 0
			}
			return s, nil

		case "k", "up":
			s.err = ""
			s.success = ""
			s.cursor--
			if s.cursor < 0 {
				s.cursor = s.totalFields - 1
			}
			return s, nil

		case "h", "left":
			s.err = ""
			s.success = ""
			switch s.cursor {
			case 1: // purchaseType
				s.purchaseTypeIdx--
				if s.purchaseTypeIdx < 0 {
					s.purchaseTypeIdx = len(s.purchaseType) - 1
				}

				return s, nil
			case 2: // timeInForce
				s.timeInForceIdx--
				if s.timeInForceIdx < 0 {
					s.timeInForceIdx = len(s.timeInForce) - 1
				}

				return s, nil
			}

		case "l", "right":
			s.err = ""
			s.success = ""
			switch s.cursor {
			case 1: // purchaseType
				s.purchaseTypeIdx++
				if s.purchaseTypeIdx >= len(s.purchaseType) {
					s.purchaseTypeIdx = 0
				}

				return s, nil
			case 2: // timeInForce
				s.timeInForceIdx++
				if s.timeInForceIdx >= len(s.timeInForce) {
					s.timeInForceIdx = 0
				}

				return s, nil
			}

		case "enter":
			s.err = ""
			s.success = ""
			if err := s.submitOrder(); err != nil {
				s.err = err.Error()
			} else {
				s.success = "Order submitted successfully!"
			}

			return s, func() tea.Msg {
				return messages.ReloadMsg{
					Page: messages.ProfilePageNumber,
				}
			}

		case "esc":
			return s, func() tea.Msg {
				return messages.SmartPageSwitchMsg{
					Page: messages.ProfilePageNumber,
				}
			}

		case "w", "W":
			return s, func() tea.Msg {
				return messages.SmartPageSwitchMsg{
					Page: messages.WatchlistPageNumber,
				}
			}

		case "i", "I":
			return s, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.TradingInfoPageNumber,
				}
			}

		default:
			key := msg.String()
			if len(key) == 1 {
				if []byte(key)[0] < '0' || []byte(key)[0] > '9' {
					return s, nil
				}
			}

			if s.cursor == 0 {
				s.quantity, cmd = s.quantity.Update(msg)
			}

			return s, nil
		}
	}

	return s, cmd
}

func (s SellPage) View() string {
	// Build header
	header := titleStyle.Render(fmt.Sprintf("📈 %s - %s Order", strings.ToUpper(s.Symbol), strings.ToUpper(s.side)))

	var fields []string
	idx := 0

	if s.cursor == idx {
		s.quantity.Focus()
	} else {
		s.quantity.Blur()
	}
	fields = append(fields, s.renderField("Quantity", s.quantity.View(), s.cursor == idx, false))
	idx++

	// Purchase Type (slider)
	fields = append(fields, s.renderField("Purchase Type", s.renderSlider(s.purchaseType, s.purchaseTypeIdx, s.cursor == idx), s.cursor == idx, true))
	idx++

	// Time In Force (slider)
	fields = append(fields, s.renderField("Time In Force", s.renderSlider(s.timeInForce, s.timeInForceIdx, s.cursor == idx), s.cursor == idx, true))
	idx++

	content := lipgloss.JoinVertical(lipgloss.Center, fields...)

	if s.err != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", errorStyle.Render("❌ "+s.err))
	}
	if s.success != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", successStyle.Render("✓ "+s.success))
	}

	help := helpStyle.Render("j/k/↑/↓: navigate | h/l/←/→: change slider | enter: submit | esc: back | w: watchlist page | i: information page | q: quit")

	headerHeight := lipgloss.Height(header)
	contentHeight := lipgloss.Height(content)
	helpHeight := lipgloss.Height(help)

	centeredHeader := lipgloss.Place(s.BaseModel.Width, headerHeight, lipgloss.Center, lipgloss.Top, header)
	centeredContent := lipgloss.Place(s.BaseModel.Width, contentHeight, lipgloss.Center, lipgloss.Top, content)
	centeredHelp := lipgloss.Place(s.BaseModel.Width, helpHeight, lipgloss.Center, lipgloss.Top, help)

	// Build final view with vertical spacing
	finalView := lipgloss.JoinVertical(
		lipgloss.Center,
		centeredHeader,
		strings.Repeat("\n", 12),
		centeredContent,
		"",
		centeredHelp,
	)

	return finalView
}

func (s SellPage) renderField(label, value string, focused, slider bool) string {
	styledLabel := labelStyle.Render(label + ":")

	var fieldStyle lipgloss.Style
	if focused {
		if slider {
			fieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Background(lipgloss.Color("#2a2a4e")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FFFF")).
				Width(30).
				Align(lipgloss.Center)
		} else {
			fieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Background(lipgloss.Color("#2a2a4e")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FFFF")).
				Width(30)
		}
	} else {
		if slider {
			fieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#666666")).
				Width(30).
				Align(lipgloss.Center)
		} else {
			fieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#666666")).
				Width(30)
		}
	}

	styledValue := fieldStyle.Render(value)

	return lipgloss.JoinVertical(lipgloss.Center, styledLabel, styledValue)
}

func (s SellPage) renderSlider(options []string, selectedIdx int, focused bool) string {
	selected := strings.ToUpper(options[selectedIdx])

	if focused {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Render("◀ " + selected + " ▶")
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Render("◀ " + selected + " ▶")
}

func (s *SellPage) submitOrder() error {
	data := make(map[string]any)

	data["symbol"] = s.Symbol
	data["side"] = s.side
	data["type"] = s.purchaseType[s.purchaseTypeIdx]
	data["time_in_force"] = s.timeInForce[s.timeInForceIdx]

	qty := strings.TrimSpace(s.quantity.Value())
	if qty == "" {
		return fmt.Errorf("quantity is required")
	}

	qty = strings.ReplaceAll(qty, ",", ".")
	dotCount := strings.Count(qty, ".")
	if dotCount > 1 {
		return errors.New("Error invalid number")
	}

	if s.purchaseType[s.purchaseTypeIdx] == "market" && s.timeInForce[s.timeInForceIdx] == "day" {
		// can be a float
		quantity, err := strconv.ParseFloat(qty, 64)
		if err != nil {
			return err
		}

		if quantity > s.MaxQuantity {
			return errors.New("Error the ammount of stock you are trying to sell is bigger than what you have")
		}

		data["qty"] = quantity
	} else {
		if dotCount > 0 {
			return errors.New("Error quantity must be an integer")
		}

		quantity, err := strconv.Atoi(qty)
		if err != nil {
			return err
		}

		if quantity > int(s.MaxQuantity) {
			return errors.New("Error the ammount of stock you are trying to sell is bigger than what you have")
		}

		data["qty"] = qty
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode request: %v", err)
	}

	_, err = requests.MakeRequest(http.MethodPost, requests.BaseURL+"/trading", bytes.NewReader(jsonData), http.DefaultClient, s.BaseModel.Token)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	return nil
}

func (s *SellPage) Reload() {
	s.cursor = 0
	s.quantity.SetValue("1")
	s.purchaseTypeIdx = 0
	s.timeInForceIdx = 0
	s.err = ""
	s.success = ""
}
