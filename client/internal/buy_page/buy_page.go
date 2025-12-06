package buypage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TakeProfit struct {
	limitPrice textinput.Model
}

type StopLoss struct {
	stopPrice  textinput.Model
	limitPrice textinput.Model
}

type BuyPage struct {
	BaseModel        basemodel.BaseModel
	Symbol           string
	quantity         textinput.Model
	side             string
	purchaseType     []string
	purchaseTypeIdx  int
	timeInForce      []string
	timeInForceIdx   int
	additionalFields map[string][]textinput.Model
	takeProfit       TakeProfit
	stopLoss         StopLoss
	cursor           int
	totalFields      int
	err              string
	success          string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB88FF")).
			Width(20)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true)

	sliderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 2).
			Margin(0, 1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Margin(1, 0)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)
)

func NewBuyPage(client *http.Client) BuyPage {
	quantity := textinput.New()
	quantity.Placeholder = "Quantity"
	quantity.Width = 8
	quantity.SetValue("1")
	quantity.CharLimit = 10

	limitPrice := textinput.New()
	limitPrice.Placeholder = "Limit price"
	limitPrice.Width = 12
	limitPrice.CharLimit = 20

	stopPrice := textinput.New()
	stopPrice.Placeholder = "Stop price"
	stopPrice.Width = 12
	stopPrice.CharLimit = 20

	trailPrice := textinput.New()
	trailPrice.Placeholder = "Trail price"
	trailPrice.Width = 12
	trailPrice.CharLimit = 20

	trailPercent := textinput.New()
	trailPercent.Placeholder = "Trail percent"
	trailPercent.Width = 12
	trailPercent.CharLimit = 20

	takeProfit := textinput.New()
	takeProfit.Placeholder = "Limit price"
	takeProfit.Width = 12
	takeProfit.CharLimit = 20

	stopLossStopPrice := textinput.New()
	stopLossStopPrice.Placeholder = "Stop price"
	stopLossStopPrice.Width = 12
	stopLossStopPrice.CharLimit = 20

	stopLossLimitPrice := textinput.New()
	stopLossLimitPrice.Placeholder = "Limit price"
	stopLossLimitPrice.Width = 12
	stopLossLimitPrice.CharLimit = 20

	stopLoss := StopLoss{
		stopPrice:  stopLossStopPrice,
		limitPrice: stopLossLimitPrice,
	}

	return BuyPage{
		BaseModel:       basemodel.BaseModel{Client: client},
		quantity:        quantity,
		side:            "buy",
		purchaseType:    []string{"market", "limit", "stop", "stop_limit", "trailing_stop"},
		purchaseTypeIdx: 0,
		timeInForce:     []string{"day", "gtc", "opg", "cls", "ioc", "fok"},
		timeInForceIdx:  0,
		additionalFields: map[string][]textinput.Model{
			"limit": {
				limitPrice,
			},
			"stop_limit": {
				limitPrice,
				stopPrice,
			},
			"stop": {
				stopPrice,
			},
			"trailing_stop": {
				trailPrice,
				trailPercent,
			},
		},
		takeProfit: TakeProfit{
			limitPrice: takeProfit,
		},
		stopLoss: stopLoss,
		cursor:   0,
	}
}

func (b BuyPage) Init() tea.Cmd {
	return textinput.Blink
}

func (b *BuyPage) calculateTotalFields() int {
	count := 3 // quantity, purchaseType, timeInForce

	if fields, ok := b.additionalFields[b.purchaseType[b.purchaseTypeIdx]]; ok {
		count += len(fields)
	}

	count += 3 // takeProfit.limitPrice, stopLoss.stopPrice, stopLoss.limitPrice

	return count
}

func (b BuyPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	b.totalFields = b.calculateTotalFields()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return b, tea.Quit

		case "j", "down":
			b.err = ""
			b.success = ""
			b.cursor++
			if b.cursor >= b.totalFields {
				b.cursor = 0
			}
			return b, nil

		case "k", "up":
			b.err = ""
			b.success = ""
			b.cursor--
			if b.cursor < 0 {
				b.cursor = b.totalFields - 1
			}
			return b, nil

		case "h", "left":
			b.err = ""
			b.success = ""
			switch b.cursor {
			case 1: // purchaseType
				b.purchaseTypeIdx--
				if b.purchaseTypeIdx < 0 {
					b.purchaseTypeIdx = len(b.purchaseType) - 1
				}

				b.totalFields = b.calculateTotalFields()
				return b, nil
			case 2: // timeInForce
				b.timeInForceIdx--
				if b.timeInForceIdx < 0 {
					b.timeInForceIdx = len(b.timeInForce) - 1
				}

				b.totalFields = b.calculateTotalFields()
				return b, nil
			}

		case "l", "right":
			b.err = ""
			b.success = ""
			switch b.cursor {
			case 1: // purchaseType
				b.purchaseTypeIdx++
				if b.purchaseTypeIdx >= len(b.purchaseType) {
					b.purchaseTypeIdx = 0
				}

				b.totalFields = b.calculateTotalFields()
				return b, nil
			case 2: // timeInForce
				b.timeInForceIdx++
				if b.timeInForceIdx >= len(b.timeInForce) {
					b.timeInForceIdx = 0
				}

				b.totalFields = b.calculateTotalFields()
				return b, nil
			}

		case "enter":
			b.err = ""
			b.success = ""
			if err := b.submitOrder(); err != nil {
				b.err = err.Error()
			} else {
				b.success = "Order submitted successfully!"
			}
			return b, nil
		}
	}

	// Update the focused input
	fieldIdx := b.getFieldIndex()
	if input := b.getInputAtIndex(fieldIdx); input != nil {
		*input, cmd = input.Update(msg)
	}

	return b, cmd
}

func (b *BuyPage) getFieldIndex() int {
	idx := 0
	if b.cursor == idx {
		return idx // quantity
	}
	idx++

	if b.cursor == idx {
		return -1 // purchaseType (slider)
	}
	idx++

	if b.cursor == idx {
		return -1 // timeInForce (slider)
	}
	idx++

	// Additional fields
	if fields, ok := b.additionalFields[b.purchaseType[b.purchaseTypeIdx]]; ok {
		for i := range fields {
			if b.cursor == idx {
				return 100 + i // offset for additional fields
			}
			idx++
		}
	}

	// Take profit
	if b.cursor == idx {
		return 200
	}
	idx++

	// Stop loss
	if b.cursor == idx {
		return 300
	}
	idx++

	if b.cursor == idx {
		return 301
	}

	return -1
}

func (b *BuyPage) getInputAtIndex(idx int) *textinput.Model {
	if idx == 0 {
		return &b.quantity
	}
	if idx >= 100 && idx < 200 {
		fieldIdx := idx - 100
		if fields, ok := b.additionalFields[b.purchaseType[b.purchaseTypeIdx]]; ok {
			if fieldIdx < len(fields) {
				return &fields[fieldIdx]
			}
		}
	}
	if idx == 200 {
		return &b.takeProfit.limitPrice
	}
	if idx == 300 {
		return &b.stopLoss.stopPrice
	}
	if idx == 301 {
		return &b.stopLoss.limitPrice
	}
	return nil
}

func (b BuyPage) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(fmt.Sprintf("📈 %s - %s Order", strings.ToUpper(b.Symbol), strings.ToUpper(b.side))))
	s.WriteString("\n\n")

	idx := 0

	// Quantity
	cursor := " "
	if b.cursor == idx {
		cursor = cursorStyle.Render(">")
		b.quantity.Focus()
	} else {
		b.quantity.Blur()
	}
	s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, labelStyle.Render("Quantity:"), inputStyle.Render(b.quantity.View())))
	idx++

	// Purchase Type (slider)
	cursor = " "
	if b.cursor == idx {
		cursor = cursorStyle.Render(">")
	}
	s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, labelStyle.Render("Purchase Type:"),
		b.renderSlider(b.purchaseType, b.purchaseTypeIdx, b.cursor == idx)))
	idx++

	// Time In Force (slider)
	cursor = " "
	if b.cursor == idx {
		cursor = cursorStyle.Render(">")
	}
	s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, labelStyle.Render("Time In Force:"),
		b.renderSlider(b.timeInForce, b.timeInForceIdx, b.cursor == idx)))
	idx++

	// Additional fields for current purchase type
	if fields, ok := b.additionalFields[b.purchaseType[b.purchaseTypeIdx]]; ok {
		s.WriteString("\n")
		for i, field := range fields {
			cursor = " "
			if b.cursor == idx {
				cursor = cursorStyle.Render(">")
				field.Focus()
			} else {
				field.Blur()
			}

			label := field.Placeholder
			// Special handling for trailing_stop (only one required)
			if b.purchaseType[b.purchaseTypeIdx] == "trailing_stop" {
				if i == 0 {
					label += " (OR)"
				} else {
					label = "  " + label
				}
			}

			s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, labelStyle.Render(label+":"),
				inputStyle.Render(field.View())))
			idx++
		}
	}

	// Take Profit (optional)
	s.WriteString("\n")
	cursor = " "
	if b.cursor == idx {
		cursor = cursorStyle.Render(">")
		b.takeProfit.limitPrice.Focus()
	} else {
		b.takeProfit.limitPrice.Blur()
	}
	s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, labelStyle.Render("Take Profit (opt):"),
		inputStyle.Render(b.takeProfit.limitPrice.View())))
	idx++

	// Stop Loss (optional)
	cursor = " "
	if b.cursor == idx {
		cursor = cursorStyle.Render(">")
		b.stopLoss.stopPrice.Focus()
	} else {
		b.stopLoss.stopPrice.Blur()
	}
	s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, labelStyle.Render("Stop Loss Stop (opt):"),
		inputStyle.Render(b.stopLoss.stopPrice.View())))
	idx++

	cursor = " "
	if b.cursor == idx {
		cursor = cursorStyle.Render(">")
		b.stopLoss.limitPrice.Focus()
	} else {
		b.stopLoss.limitPrice.Blur()
	}
	s.WriteString(fmt.Sprintf("%s %s %s\n", cursor, labelStyle.Render("Stop Loss Limit (opt):"),
		inputStyle.Render(b.stopLoss.limitPrice.View())))

	// Error/Success messages
	if b.err != "" {
		s.WriteString("\n\n")
		s.WriteString(errorStyle.Render("❌ Error: " + b.err))
	}
	if b.success != "" {
		s.WriteString("\n\n")
		s.WriteString(successStyle.Render("✓ " + b.success))
	}

	// Help text
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("j/k/↑/↓: navigate | h/l/←/→: change slider | enter: submit | esc: quit"))

	return s.String()
}

func (b BuyPage) renderSlider(options []string, selectedIdx int, focused bool) string {
	var parts []string

	parts = append(parts, "◀")

	for i, opt := range options {
		if i == selectedIdx {
			if focused {
				parts = append(parts, selectedStyle.Render("["+strings.ToUpper(opt)+"]"))
			} else {
				parts = append(parts, "["+strings.ToUpper(opt)+"]")
			}
		} else {
			parts = append(parts, strings.ToUpper(opt))
		}
	}

	parts = append(parts, "▶")

	style := sliderStyle
	if focused {
		style = style.Background(lipgloss.Color("#2a2a4e"))
	}

	return style.Render(strings.Join(parts, " "))
}

func (b *BuyPage) submitOrder() error {
	// Validate and build request
	data := make(map[string]any)

	data["symbol"] = b.Symbol
	data["side"] = b.side
	data["type"] = b.purchaseType[b.purchaseTypeIdx]
	data["time_in_force"] = b.timeInForce[b.timeInForceIdx]

	qty := strings.TrimSpace(b.quantity.Value())
	if qty == "" {
		return fmt.Errorf("quantity is required")
	}

	qty = strings.ReplaceAll(qty, ",", ".")
	dotCount := strings.Count(qty, ".")
	if dotCount > 1 {
		return errors.New("Error invalid number")
	}

	if b.purchaseType[b.purchaseTypeIdx] == "market" && b.timeInForce[b.timeInForceIdx] == "day" {
		// can be a float
		data["qty"] = qty
	} else {
		if dotCount > 0 {
			return errors.New("Error quantity must be an integer")
		}

		data["qty"] = qty
	}

	currentType := b.purchaseType[b.purchaseTypeIdx]
	if fields, ok := b.additionalFields[currentType]; ok {
		if currentType == "trailing_stop" {
			// Only one of trail_price or trail_percent is required
			trailPrice := strings.TrimSpace(fields[0].Value())
			trailPercent := strings.TrimSpace(fields[1].Value())

			if trailPrice == "" && trailPercent == "" {
				return errors.New("Error either trail price or trail percent is required")
			}

			if trailPrice != "" {
				if strings.Count(trailPrice, ".") > 1 {
					return errors.New("Error invalid trail price number")
				}

				data["trail_price"] = trailPrice
			}

			if trailPercent != "" {
				if strings.Count(trailPercent, ".") > 1 {
					return errors.New("Error invalid trail price number")
				}

				data["trail_percent"] = trailPercent
			}
		} else {
			fieldNames := b.getFieldNames(currentType)
			for i, field := range fields {
				val := strings.TrimSpace(field.Value())
				if val == "" {
					return fmt.Errorf("Error %s is required", field.Placeholder)
				}

				data[fieldNames[i]] = val
			}
		}
	}

	// Add take profit if provided
	if tp := strings.TrimSpace(b.takeProfit.limitPrice.Value()); tp != "" {
		if strings.Count(tp, ".") > 1 {
			return errors.New("Error invalid take profit value")
		}

		data["take_profit"] = map[string]string{"limit_price": tp}
	}

	// Add stop loss if provided
	slStop := strings.TrimSpace(b.stopLoss.stopPrice.Value())
	slLimit := strings.TrimSpace(b.stopLoss.limitPrice.Value())

	if slStop != "" && slLimit != "" {
		stopLoss := make(map[string]string)

		if strings.Count(slStop, ".") > 1 {
			return errors.New("Error invalid stop loss stop price")
		}

		stopLoss["stop_price"] = slStop

		if strings.Count(slLimit, ".") > 1 {
			return errors.New("Error invalid stop loss limit price")
		}

		stopLoss["limit_price"] = slLimit

		data["stop_loss"] = stopLoss
	} else {
		return errors.New("Error both stop price and stop limit are required if you want to have a stop loss")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode request: %v", err)
	}

	body, err := requests.MakeRequest(http.MethodPost, requests.BaseURL+"/trading", bytes.NewReader(jsonData), http.DefaultClient, b.BaseModel.Token)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	log.Println(string(body))

	return nil
}

func (b *BuyPage) getFieldNames(purchaseType string) []string {
	switch purchaseType {
	case "limit":
		return []string{"limit_price"}
	case "stop":
		return []string{"stop_price"}
	case "stop_limit":
		return []string{"limit_price", "stop_price"}
	case "trailing_stop":
		return []string{"trail_price", "trail_percent"}
	default:
		return []string{}
	}
}
