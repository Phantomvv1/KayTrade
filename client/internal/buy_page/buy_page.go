package buypage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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

func NewBuyPage(client *http.Client) BuyPage {
	quantity := textinput.New()
	quantity.Placeholder = "Quantity"
	quantity.Width = 28
	quantity.SetValue("1")
	quantity.CharLimit = 10

	limitPrice := textinput.New()
	limitPrice.Placeholder = "Limit price"
	limitPrice.Width = 28
	limitPrice.CharLimit = 20

	stopPrice := textinput.New()
	stopPrice.Placeholder = "Stop price"
	stopPrice.Width = 28
	stopPrice.CharLimit = 20

	trailPrice := textinput.New()
	trailPrice.Placeholder = "Trail price"
	trailPrice.Width = 28
	trailPrice.CharLimit = 20

	trailPercent := textinput.New()
	trailPercent.Placeholder = "Trail percent"
	trailPercent.Width = 28
	trailPercent.CharLimit = 20

	takeProfit := textinput.New()
	takeProfit.Placeholder = "Limit price"
	takeProfit.Width = 28
	takeProfit.CharLimit = 20

	stopLossStopPrice := textinput.New()
	stopLossStopPrice.Placeholder = "Stop price"
	stopLossStopPrice.Width = 28
	stopLossStopPrice.CharLimit = 20

	stopLossLimitPrice := textinput.New()
	stopLossLimitPrice.Placeholder = "Limit price"
	stopLossLimitPrice.Width = 28
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

		default:
			fieldIdx := b.getFieldIndex()
			if input := b.getInputAtIndex(fieldIdx); input != nil {
				input.Focus()
				updatedInput, cmd := input.Update(msg)
				b.setInputAtIndex(fieldIdx, updatedInput)
				input.Blur()
				return b, cmd
			}

			return b, nil
		}
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

func (b *BuyPage) setInputAtIndex(idx int, input textinput.Model) {
	if idx == 0 {
		b.quantity = input
		return
	}
	if idx >= 100 && idx < 200 {
		fieldIdx := idx - 100
		if fields, ok := b.additionalFields[b.purchaseType[b.purchaseTypeIdx]]; ok {
			if fieldIdx < len(fields) {
				fields[fieldIdx] = input
				b.additionalFields[b.purchaseType[b.purchaseTypeIdx]] = fields
			}
		}
		return
	}
	if idx == 200 {
		b.takeProfit.limitPrice = input
		return
	}
	if idx == 300 {
		b.stopLoss.stopPrice = input
		return
	}
	if idx == 301 {
		b.stopLoss.limitPrice = input
		return
	}
}

func (b BuyPage) View() string {
	// Build header
	header := titleStyle.Render(fmt.Sprintf("📈 %s - %s Order", strings.ToUpper(b.Symbol), strings.ToUpper(b.side)))

	// Build form fields as a slice
	var fields []string
	idx := 0

	// Quantity
	if b.cursor == idx {
		b.quantity.Focus()
	} else {
		b.quantity.Blur()
	}
	fields = append(fields, b.renderField("Quantity", b.quantity.View(), b.cursor == idx, false))
	idx++

	// Purchase Type (slider)
	fields = append(fields, b.renderField("Purchase Type", b.renderSlider(b.purchaseType, b.purchaseTypeIdx, b.cursor == idx), b.cursor == idx, true))
	idx++

	// Time In Force (slider)
	fields = append(fields, b.renderField("Time In Force", b.renderSlider(b.timeInForce, b.timeInForceIdx, b.cursor == idx), b.cursor == idx, true))
	idx++

	// Additional fields for current purchase type
	currentType := b.purchaseType[b.purchaseTypeIdx]
	if additionalFields, ok := b.additionalFields[currentType]; ok {
		for i := range additionalFields {
			field := &additionalFields[i]
			if b.cursor == idx {
				field.Focus()
			} else {
				field.Blur()
			}

			label := field.Placeholder
			// Special handling for trailing_stop (only one required)
			if currentType == "trailing_stop" {
				if i == 0 {
					label += " (OR)"
				} else {
					label = "  " + label
				}
			}

			fields = append(fields, b.renderField(label, field.View(), b.cursor == idx, false))
			idx++
		}
	}

	// Take Profit (optional)
	if b.cursor == idx {
		b.takeProfit.limitPrice.Focus()
	} else {
		b.takeProfit.limitPrice.Blur()
	}
	fields = append(fields, b.renderField("Take Profit (opt)", b.takeProfit.limitPrice.View(), b.cursor == idx, false))
	idx++

	// Stop Loss (optional)
	if b.cursor == idx {
		b.stopLoss.stopPrice.Focus()
	} else {
		b.stopLoss.stopPrice.Blur()
	}
	fields = append(fields, b.renderField("Stop Loss Stop (opt)", b.stopLoss.stopPrice.View(), b.cursor == idx, false))
	idx++

	if b.cursor == idx {
		b.stopLoss.limitPrice.Focus()
	} else {
		b.stopLoss.limitPrice.Blur()
	}
	fields = append(fields, b.renderField("Stop Loss Limit (opt)", b.stopLoss.limitPrice.View(), b.cursor == idx, false))

	content := lipgloss.JoinVertical(lipgloss.Center, fields...)

	// Add error/success messages if present
	if b.err != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", errorStyle.Render("❌ "+b.err))
	}
	if b.success != "" {
		content = lipgloss.JoinVertical(lipgloss.Center, content, "", successStyle.Render("✓ "+b.success))
	}

	help := helpStyle.Render("j/k/↑/↓: navigate | h/l/←/→: change slider | enter: submit | q: quit")

	// Calculate vertical spacing
	headerHeight := lipgloss.Height(header)
	contentHeight := lipgloss.Height(content)
	helpHeight := lipgloss.Height(help)

	// Center everything horizontally
	centeredHeader := lipgloss.Place(b.BaseModel.Width, headerHeight, lipgloss.Center, lipgloss.Top, header)
	centeredContent := lipgloss.Place(b.BaseModel.Width, contentHeight, lipgloss.Center, lipgloss.Top, content)
	centeredHelp := lipgloss.Place(b.BaseModel.Width, helpHeight, lipgloss.Center, lipgloss.Top, help)

	// Build final view with vertical spacing
	finalView := lipgloss.JoinVertical(
		lipgloss.Center,
		centeredHeader,
		strings.Repeat("\n", 16-b.calculateTotalFields()),
		centeredContent,
		"",
		centeredHelp,
	)

	return finalView
}

func (b BuyPage) renderField(label, value string, focused, slider bool) string {
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

func (b BuyPage) renderSlider(options []string, selectedIdx int, focused bool) string {
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

func (b *BuyPage) Reload() {
	b.cursor = 0
	b.quantity.SetValue("1")
	b.purchaseTypeIdx = 0
	b.timeInForceIdx = 0
	for _, fields := range b.additionalFields {
		for _, field := range fields {
			field.SetValue("")
		}
	}
	b.takeProfit.limitPrice.SetValue("")
	b.stopLoss.stopPrice.SetValue("")
	b.stopLoss.limitPrice.SetValue("")
	b.err = ""
	b.success = ""
}
