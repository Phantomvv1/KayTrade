package tradinginfopage

import (
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TradingInfoPage struct {
	BaseModel                       basemodel.BaseModel
	Viewport                        viewport.Model
	ready                           bool
	qunatityExplanation             string
	timeInForceExplanation          string
	limitPriceExplanation           string
	stopPriceExplanation            string
	trailPriceAndPercentExplanation string
	takeProfitExplanation           string
	stopLossExplanation             string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1)

	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#BB88FF")).
				Bold(true).
				Underline(true).
				MarginTop(1).
				MarginBottom(1)

	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			PaddingLeft(2).
			Width(80)

	subItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			PaddingLeft(4).
			Width(76)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(1, 0)
)

func NewTradingInfoPage() TradingInfoPage {
	return TradingInfoPage{
		qunatityExplanation: `For equities, the number of shares to trade. Can be fractionable for only market and day order types.
For Fixed Income securities, qty represents the order size in par value (face value). For example, to place an order for 1 bond with a face value of $1,000, provide a qty of 1000.`,
		timeInForceExplanation: `day: A day order is eligible for execution only on the day it is live. By default, the order is only valid during Regular Trading Hours (9:30am - 4:00pm ET). If unfilled after the closing auction, it is automatically canceled. If submitted after the close, it is queued and submitted the following trading day. However, if marked as eligible for extended hours, the order can also execute during supported extended hours.

gtc: The order is good until canceled. Non-marketable GTC limit orders are subject to price adjustments to offset corporate actions affecting the issue. We do not currently support Do Not Reduce(DNR) orders to opt out of such price adjustments.

opg: Use this TIF with a market/limit order type to submit "market on open" (MOO) and "limit on open" (LOO) orders. This order is eligible to execute only in the market opening auction. Any unfilled orders after the open will be cancelled. OPG orders submitted after 9:28am but before 7:00pm ET will be rejected. OPG orders submitted after 7:00pm will be queued and routed to the following day's opening auction. On open/on close orders are routed to the primary exchange. Such orders do not necessarily execute exactly at 9:30am / 4:00pm ET but execute per the exchange's auction rules.

cls: Use this TIF with a market/limit order type to submit "market on close" (MOC) and "limit on close" (LOC) orders. This order is eligible to execute only in the market closing auction. Any unfilled orders after the close will be cancelled. CLS orders submitted after 3:50pm but before 7:00pm ET will be rejected. CLS orders submitted after 7:00pm will be queued and routed to the following day's closing auction.

ioc: An Immediate Or Cancel (IOC) order requires all or part of the order to be executed immediately. Any unfilled portion of the order is canceled. Most market makers who receive IOC orders will attempt to fill the order on a principal basis only, and cancel any unfilled balance. On occasion, this can result in the entire order being cancelled if the market maker does not have any existing inventory of the security in question.

fok: A Fill or Kill (FOK) order is only executed if the entire order quantity can be filled, otherwise the order is canceled.`,
		limitPriceExplanation: `Required if type is limit or stop_limit.
The price is expressed in percentage of par value (face value). Price is always clean price, meaning it does not include accrued interest.`,
		stopPriceExplanation:            "Required if type is stop or stop_limit",
		trailPriceAndPercentExplanation: "If type is trailing_stop, then one of trail_price or trail_percent is required",
		takeProfitExplanation:           "Takes in a number value for limit_price",
		stopLossExplanation:             "Takes in number values for stop_price and limit_price",
		Viewport:                        viewport.New(130, 40),
		ready:                           false,
	}
}

func (t TradingInfoPage) Init() tea.Cmd {
	return nil
}

func (t TradingInfoPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if !t.ready {
		t.ready = true
		t.Viewport.SetContent(t.buildContent())
		t.Viewport, cmd = t.Viewport.Update(msg)
		return t, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return t, tea.Quit

		case "esc":
			return t, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.BuyPageNumber,
				}
			}

		}
	}

	t.Viewport, cmd = t.Viewport.Update(msg)
	return t, cmd
}

func (t TradingInfoPage) View() string {
	title := titleStyle.Render("📊 Trading Order Information Guide")
	help := helpStyle.Render("↑/↓/k/j: scroll • esc: back • q/ctrl+c: quit")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.Place(t.BaseModel.Width, 1, lipgloss.Center, lipgloss.Top, title),
		t.Viewport.View(),
		lipgloss.Place(t.BaseModel.Width, 1, lipgloss.Center, lipgloss.Top, help),
	)
}

func (t TradingInfoPage) buildContent() string {
	var sections []string

	// Quantity Section
	sections = append(sections,
		sectionTitleStyle.Render("📦 Quantity"),
		contentStyle.Render(t.qunatityExplanation),
	)

	// Time In Force Section
	sections = append(sections,
		"",
		sectionTitleStyle.Render("⏰ Time In Force"),
	)

	// Parse and format Time In Force items
	tifItems := strings.Split(t.timeInForceExplanation, "\n\n")
	for _, item := range tifItems {
		if strings.TrimSpace(item) != "" {
			sections = append(sections, subItemStyle.Render(strings.TrimSpace(item)))
		}
	}

	// Limit Price Section
	sections = append(sections,
		"",
		sectionTitleStyle.Render("💰 Limit Price"),
		contentStyle.Render(strings.TrimSpace(t.limitPriceExplanation)),
	)

	// Stop Price Section
	sections = append(sections,
		"",
		sectionTitleStyle.Render("🛑 Stop Price"),
		contentStyle.Render(t.stopPriceExplanation),
	)

	// Trail Price/Percent Section
	sections = append(sections,
		"",
		sectionTitleStyle.Render("📈 Trail Price & Trail Percent"),
		contentStyle.Render(t.trailPriceAndPercentExplanation),
	)

	// Take Profit Section
	sections = append(sections,
		"",
		sectionTitleStyle.Render("🎯 Take Profit"),
		contentStyle.Render(t.takeProfitExplanation),
	)

	// Stop Loss Section
	sections = append(sections,
		"",
		sectionTitleStyle.Render("🔻 Stop Loss"),
		contentStyle.Render(t.stopLossExplanation),
	)

	sections = append(sections, "", "") // Add some bottom padding

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
