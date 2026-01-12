package orderpage

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type OrderPage struct {
	BaseModel basemodel.BaseModel
	Order     *messages.Order
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#BB88FF")).
				Bold(true).
				Underline(true).
				MarginTop(1).
				MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB88FF")).
			Width(20)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true)

	statusPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500")).
				Bold(true)

	statusFilledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)

	statusCanceledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	statusExpiredStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BB88FF")).
			Padding(1, 2).
			Width(70)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	sideBuyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	sideSellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

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
			return o, func() tea.Msg {
				return messages.QuitMsg{}
			}

		case "esc":
			return o, func() tea.Msg {
				return messages.SmartPageSwitchMsg{
					Page: messages.ProfilePageNumber,
				}
			}
		}
	}
	return o, nil
}

func (o OrderPage) View() string {
	if o.Order == nil {
		return lipgloss.Place(
			o.BaseModel.Width,
			o.BaseModel.Height,
			lipgloss.Center,
			lipgloss.Center,
			"No order selected",
		)
	}

	sideText := ""
	if strings.ToUpper(o.Order.Side) == "BUY" {
		sideText = sideBuyStyle.Render("BUY")
	} else {
		sideText = sideSellStyle.Render("SELL")
	}

	header := titleStyle.Render(
		fmt.Sprintf("📋 Order Details — %s %s %s",
			o.Order.Quantity,
			sideText,
			strings.ToUpper(o.Order.Symbol),
		),
	)

	var sections []string

	sections = append(sections, o.renderBasicInfo())

	sections = append(sections, o.renderPricingInfo())

	sections = append(sections, o.renderStatusInfo())

	sections = append(sections, o.renderAdditionalInfo())

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	centeredContent := lipgloss.Place(o.BaseModel.Width, lipgloss.Height(content), lipgloss.Center, lipgloss.Top, content)

	help := helpStyle.Render("esc: back • q: quit")
	centeredHelp := lipgloss.Place(o.BaseModel.Width, lipgloss.Height(help), lipgloss.Center, lipgloss.Top, help)

	centeredHeader := lipgloss.Place(o.BaseModel.Width, lipgloss.Height(header), lipgloss.Center, lipgloss.Top, header)

	finalView := lipgloss.JoinVertical(
		lipgloss.Center,
		"\n",
		centeredHeader,
		centeredContent,
		centeredHelp,
	)

	return lipgloss.Place(
		o.BaseModel.Width,
		o.BaseModel.Height,
		lipgloss.Center,
		lipgloss.Top,
		finalView,
	)
}

func (o OrderPage) renderBasicInfo() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("📊 Basic Information"))
	rows = append(rows, o.renderField("Order ID", o.Order.ID))
	rows = append(rows, o.renderField("Symbol", strings.ToUpper(o.Order.Symbol)))
	rows = append(rows, o.renderField("Quantity", o.Order.Quantity))

	side := strings.ToUpper(o.Order.Side)
	if side == "BUY" {
		rows = append(rows, labelStyle.Render("Side:")+"  "+sideBuyStyle.Render(side))
	} else {
		rows = append(rows, labelStyle.Render("Side:")+"  "+sideSellStyle.Render(side))
	}

	rows = append(rows, o.renderField("Order Type", strings.ToUpper(o.Order.OrderType)))
	rows = append(rows, o.renderField("Time In Force", strings.ToUpper(o.Order.TimeInForce)))
	rows = append(rows, o.renderField("Asset Class", strings.ToUpper(o.Order.AssetClass)))

	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (o OrderPage) renderPricingInfo() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("💰 Pricing Information"))

	if o.Order.LimitPrice != nil && *o.Order.LimitPrice != "" {
		rows = append(rows, o.renderField("Limit Price", "$"+*o.Order.LimitPrice))
	}

	if o.Order.StopPrice != nil && *o.Order.StopPrice != "" {
		rows = append(rows, o.renderField("Stop Price", "$"+*o.Order.StopPrice))
	}

	if o.Order.TrailPrice != nil && *o.Order.TrailPrice != "" {
		rows = append(rows, o.renderField("Trail Price", "$"+*o.Order.TrailPrice))
	}

	if o.Order.TrailPercent != nil && *o.Order.TrailPercent != "" {
		rows = append(rows, o.renderField("Trail Percent", *o.Order.TrailPercent+"%"))
	}

	if o.Order.FilledAvgPrice != "" {
		rows = append(rows, o.renderField("Filled Avg Price", "$"+o.Order.FilledAvgPrice))
	}

	if o.Order.FilledQty != "" && o.Order.FilledQty != "0" {
		rows = append(rows, o.renderField("Filled Quantity", o.Order.FilledQty))
	}

	if o.Order.Notional != nil && *o.Order.Notional != "" {
		rows = append(rows, o.renderField("Notional Value", "$"+*o.Order.Notional))
	}

	// Only show this section if there's pricing info
	if len(rows) > 1 {
		return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
	}
	return ""
}

func (o OrderPage) renderStatusInfo() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("⏱️  Status & Timing"))

	status := strings.ToUpper(o.Order.Status)
	var statusStyled string
	switch status {
	case "FILLED", "COMPLETED":
		statusStyled = statusFilledStyle.Render(status)
	case "CANCELED":
		statusStyled = statusCanceledStyle.Render(status)
	case "EXPIRED":
		statusStyled = statusExpiredStyle.Render(status)
	case "PENDING", "NEW", "ACCEPTED", "PENDING_NEW":
		statusStyled = statusPendingStyle.Render(status)
	default:
		statusStyled = valueStyle.Render(status)
	}
	rows = append(rows, labelStyle.Render("Status:")+"  "+statusStyled)

	if o.Order.CreatedAt != "" {
		rows = append(rows, o.renderField("Created", o.formatTimestamp(o.Order.CreatedAt)))
	}

	if o.Order.SubmittedAt != "" {
		rows = append(rows, o.renderField("Submitted", o.formatTimestamp(o.Order.SubmittedAt)))
	}

	if o.Order.UpdatedAt != "" {
		rows = append(rows, o.renderField("Updated", o.formatTimestamp(o.Order.UpdatedAt)))
	}

	if o.Order.FilledAt != "" {
		rows = append(rows, o.renderField("Filled", o.formatTimestamp(o.Order.FilledAt)))
	}

	if o.Order.CanceledAt != "" {
		rows = append(rows, o.renderField("Canceled", o.formatTimestamp(o.Order.CanceledAt)))
	}

	if o.Order.ExpiredAt != "" {
		rows = append(rows, o.renderField("Expired", o.formatTimestamp(o.Order.ExpiredAt)))
	}

	if o.Order.ExpiresAt != "" {
		rows = append(rows, o.renderField("Expires", o.formatTimestamp(o.Order.ExpiresAt)))
	}

	if o.Order.FailedAt != "" {
		rows = append(rows, o.renderField("Failed", o.formatTimestamp(o.Order.FailedAt)))
	}

	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (o OrderPage) renderAdditionalInfo() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("ℹ️  Additional Details"))

	if o.Order.AssetID != "" {
		rows = append(rows, o.renderField("Asset ID", o.Order.AssetID))
	}

	if o.Order.PositionIntent != "" {
		rows = append(rows, o.renderField("Position Intent", strings.ToUpper(o.Order.PositionIntent)))
	}

	if o.Order.Type != "" && o.Order.Type != o.Order.OrderType {
		rows = append(rows, o.renderField("Type", strings.ToUpper(o.Order.Type)))
	}

	// Only show this section if there's additional info
	if len(rows) > 1 {
		return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
	}
	return ""
}

func (o OrderPage) renderField(label, value string) string {
	return labelStyle.Render(label+":") + "  " + valueStyle.Render(value)
}

func (o OrderPage) formatTimestamp(timestamp string) string {
	// Try to parse RFC3339 format
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// If parsing fails, return original
		return timestamp
	}

	// Format as a readable string
	return t.Format("Jan 02, 2006 at 3:04 PM MST")
}
