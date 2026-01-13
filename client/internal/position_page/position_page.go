package positionpage

import (
	"fmt"
	"net/http"
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PositionPage struct {
	BaseModel basemodel.BaseModel
	Position  *messages.Position
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
			Width(30)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true)

	positiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	negativeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BB88FF")).
			Padding(1, 2).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))
)

func NewPositionPage(client *http.Client) PositionPage {
	return PositionPage{
		BaseModel: basemodel.BaseModel{Client: client},
	}
}

func (p PositionPage) Init() tea.Cmd {
	return nil
}

func (p PositionPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return p, func() tea.Msg {
				return messages.QuitMsg{}
			}

		case "esc":
			return p, func() tea.Msg {
				return messages.SmartPageSwitchMsg{
					Page: messages.ProfilePageNumber,
				}
			}
		}
	}

	return p, nil
}

func (p PositionPage) View() string {
	if p.Position == nil {
		return lipgloss.Place(
			p.BaseModel.Width,
			p.BaseModel.Height,
			lipgloss.Center,
			lipgloss.Center,
			"No position data available",
		)
	}

	header := titleStyle.Render(fmt.Sprintf("📊 Position: %s", strings.ToUpper(p.Position.Symbol)))

	basicInfo := p.renderBasicInfo()
	priceInfo := p.renderPriceInfo()
	performanceInfo := p.renderPerformanceInfo()

	leftColumn := lipgloss.JoinVertical(lipgloss.Left, basicInfo, performanceInfo)

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		lipgloss.NewStyle().MarginLeft(3).Render(priceInfo),
	)

	help := helpStyle.Render("esc: back • q: quit")

	centeredHeader := lipgloss.Place(p.BaseModel.Width, lipgloss.Height(header), lipgloss.Center, lipgloss.Top, header)

	finalView := lipgloss.JoinVertical(
		lipgloss.Center,
		"\n",
		centeredHeader,
		"",
		content,
	)

	finalWithHelp := lipgloss.JoinVertical(
		lipgloss.Center,
		finalView,
		"",
		help,
		// lipgloss.Place(p.BaseModel.Width, 1, lipgloss.Center, lipgloss.Top, help),
	)

	return lipgloss.Place(
		p.BaseModel.Width,
		p.BaseModel.Height,
		lipgloss.Center,
		lipgloss.Top,
		finalWithHelp,
	)
}

func (p PositionPage) renderBasicInfo() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("📋 Basic Information"))

	rows = append(rows, p.renderField("Symbol", strings.ToUpper(p.Position.Symbol)))
	rows = append(rows, p.renderField("Asset Class", strings.ToUpper(p.Position.AssetClass)))
	rows = append(rows, p.renderField("Exchange", strings.ToUpper(p.Position.Exchange)))
	rows = append(rows, p.renderField("Side", strings.ToUpper(p.Position.Side)))

	marginable := "No"
	if p.Position.AssetMarginable {
		marginable = "Yes"
	}
	rows = append(rows, p.renderField("Marginable", marginable))

	rows = append(rows, "")
	rows = append(rows, sectionTitleStyle.Render("📦 Quantity"))
	rows = append(rows, p.renderField("Total Quantity", p.Position.Qty))
	rows = append(rows, p.renderField("Available Quantity", p.Position.QtyAvailable))

	return boxStyle.Width(50).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (p PositionPage) renderPriceInfo() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("💰 Price Information"))

	rows = append(rows, p.renderField("Current Price", "$"+p.Position.CurrentPrice))
	rows = append(rows, p.renderField("Average Entry Price", "$"+p.Position.AvgEntryPrice))
	rows = append(rows, p.renderField("Last Day Price", "$"+p.Position.LastdayPrice))
	rows = append(rows, p.renderField("Cost Basis", "$"+p.Position.CostBasis))
	rows = append(rows, p.renderField("Market Value", "$"+p.Position.MarketValue))

	changeStr := p.Position.ChangeToday
	if strings.HasPrefix(changeStr, "-") {
		rows = append(rows, labelStyle.Render("Change Today:")+"  "+negativeStyle.Render("$"+changeStr))
	} else {
		if !strings.HasPrefix(changeStr, "+") && changeStr != "0" && changeStr != "" {
			changeStr = "+" + changeStr
		}
		rows = append(rows, labelStyle.Render("Change Today:")+"  "+positiveStyle.Render("$"+changeStr))
	}

	return boxStyle.Width(50).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (p PositionPage) renderPerformanceInfo() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("📈 Performance"))

	pl := p.Position.UnrealizedPL
	plpc := p.Position.UnrealizedPLPC

	if strings.HasPrefix(pl, "-") {
		rows = append(rows, labelStyle.Render("Unrealized P&L:")+"  "+negativeStyle.Render("$"+pl))
	} else {
		if !strings.HasPrefix(pl, "+") && pl != "0" && pl != "" {
			pl = "+" + pl
		}
		rows = append(rows, labelStyle.Render("Unrealized P&L:")+"  "+positiveStyle.Render("$"+pl))
	}

	if strings.HasPrefix(plpc, "-") {
		rows = append(rows, labelStyle.Render("Unrealized P&L %:")+"  "+negativeStyle.Render(plpc+"%"))
	} else {
		if !strings.HasPrefix(plpc, "+") && plpc != "0" && plpc != "" {
			plpc = "+" + plpc
		}
		rows = append(rows, labelStyle.Render("Unrealized P&L %:")+"  "+positiveStyle.Render(plpc+"%"))
	}

	rows = append(rows, "")
	rows = append(rows, sectionTitleStyle.Render("📊 Intraday Performance"))

	intradayPL := p.Position.UnrealizedIntradayPL
	intradayPLPC := p.Position.UnrealizedIntradayPLPC

	if strings.HasPrefix(intradayPL, "-") {
		rows = append(rows, labelStyle.Render("Intraday P&L:")+"  "+negativeStyle.Render("$"+intradayPL))
	} else {
		if !strings.HasPrefix(intradayPL, "+") && intradayPL != "0" && intradayPL != "" {
			intradayPL = "+" + intradayPL
		}
		rows = append(rows, labelStyle.Render("Intraday P&L:")+"  "+positiveStyle.Render("$"+intradayPL))
	}

	if strings.HasPrefix(intradayPLPC, "-") {
		rows = append(rows, labelStyle.Render("Intraday P&L %:")+"  "+negativeStyle.Render(intradayPLPC+"%"))
	} else {
		if !strings.HasPrefix(intradayPLPC, "+") && intradayPLPC != "0" && intradayPLPC != "" {
			intradayPLPC = "+" + intradayPLPC
		}
		rows = append(rows, labelStyle.Render("Intraday P&L %:")+"  "+positiveStyle.Render(intradayPLPC+"%"))
	}

	return boxStyle.Width(50).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (p PositionPage) renderField(label, value string) string {
	return labelStyle.Render(label+":") + "  " + valueStyle.Render(value)
}
