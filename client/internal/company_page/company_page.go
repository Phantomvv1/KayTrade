package companypage

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

const (
	tabOverview = iota
	tabHistory
	tabPrice
	tabChart
	tabLiveUpdate
	tabWarning
)

var tabTitles = []string{
	"Overview",
	"History",
	"Price",
	"Price Chart",
	"Live Price",
	"Warning",
}

type TimeFrame int

const (
	TimeFrameMinute TimeFrame = iota
	TimeFrameHour
	TimeFrameDay
	TimeFrameWeek
	TimeFrameMonth
)

var timeFrameStrings = []string{"1T", "1H", "1D", "1W", "1M"}
var timeFrameLabels = []string{"Minute", "Hour", "Day", "Week", "Month"}

type BarData struct {
	Timestamp time.Time `json:"t"`
	Open      float64   `json:"o"`
	High      float64   `json:"h"`
	Low       float64   `json:"l"`
	Close     float64   `json:"c"`
	Volume    float64   `json:"v"`
}

type BarsResponse struct {
	Bars map[string][]BarData `json:"bars"`
}

type WebSocketMsg struct {
	Type   string  `json:"T"`
	Symbol string  `json:"S"`
	Price  float64 `json:"p"`
	Size   float64 `json:"s"`
	Time   string  `json:"t"`
}

type CompanyPage struct {
	BaseModel   basemodel.BaseModel
	CompanyInfo *messages.CompanyInfo
	activeTab   int
	tabs        []int
	PrevPage    int

	// chart state
	chart     timeserieslinechart.Model
	timeFrame TimeFrame
	chartData []BarData
	// viewStart    int
	// viewEnd      int
	// zoomLevel    int
	chartLoading bool
	chartError   string

	// Live chart state
	liveChart     timeserieslinechart.Model
	liveData      []timeserieslinechart.TimePoint
	ws            *websocket.Conn
	liveConnected bool
	lastPrice     float64
	lastChange    float64
	high          float64
	low           float64
	volume        float64
	liveError     string
}

type fetchDataMsg struct {
	data []BarData
	err  error
}

type wsDataMsg struct {
	data WebSocketMsg
}

type wsErrorMsg struct {
	err error
}

type wsConnectedMsg struct{}

type addCompanyMsg struct {
	err error
}

func NewCompanyPage(client *http.Client) CompanyPage {
	// Create historical chart
	chart := timeserieslinechart.New(120, 30,
		timeserieslinechart.WithStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		))

	// Create live chart
	liveChart := timeserieslinechart.New(120, 30,
		timeserieslinechart.WithStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		))

	tabs := []int{tabOverview, tabHistory, tabPrice, tabChart}
	now := time.Now().UTC()
	check := false
	if now.Hour() < 13 || now.Hour() >= 20 { // market opens at 13:30 UTC and closes at 20:00 UTC
		check = true
	}
	if now.Hour() == 13 && now.Minute() < 30 {
		check = false
	} else if !check {
		tabs = append(tabs, tabLiveUpdate)
	}

	return CompanyPage{
		BaseModel: basemodel.BaseModel{Client: client},
		activeTab: tabOverview,
		tabs:      tabs,
		chart:     chart,
		chartData: make([]BarData, 0),
		// zoomLevel: 100,
		liveChart: liveChart,
		timeFrame: TimeFrameDay,
		liveData:  make([]timeserieslinechart.TimePoint, 0),
	}
}

func (c *CompanyPage) redrawChart() {
	if len(c.chartData) == 0 {
		return
	}

	// start := c.viewStart
	// end := c.viewEnd
	// if end > len(c.chartData) {
	// 	end = len(c.chartData)
	// }
	// if start < 0 {
	// 	start = 0
	// }
	//
	// visibleData := c.chartData[start:end]
	// if len(visibleData) == 0 {
	// 	return
	// }

	c.chart = timeserieslinechart.New(120, 30,
		timeserieslinechart.WithStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		),
		timeserieslinechart.WithXYSteps(8, 6),
	)

	for _, bar := range c.chartData {
		openPrice, highPrice, lowPrice, closePrice := c.getDataFromBar(bar)
		c.chart.Push(timeserieslinechart.TimePoint{Time: bar.Timestamp, Value: bar.Close})
		c.chart.DrawCandle(openPrice, highPrice, lowPrice, closePrice,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#7C0A02")), lipgloss.NewStyle().Foreground(lipgloss.Color("#0B6623")))
	}

	c.chart.Draw()
	// c.chart.DrawBraille()
}

func (c CompanyPage) Init() tea.Cmd {
	return nil
}

func (c CompanyPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if c.CompanyInfo != nil && c.CompanyInfo.IsNSFW && c.tabs[len(c.tabs)-1] != tabWarning {
		c.tabs = append(c.tabs, tabWarning)
	}

	switch msg := msg.(type) {
	case fetchDataMsg:
		c.chartLoading = false
		if msg.err != nil {
			return c, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.ErrorPageNumber,
					Err:  msg.err,
				}
			}
		}

		c.chartData = msg.data
		c.redrawChart()

		return c, nil

	case wsConnectedMsg:
		c.liveConnected = true
		return c, nil

	case wsDataMsg:
		err := c.processWebSocketData(msg.data)
		if err != nil {
			c.liveError = err.Error()
		}

		return c, c.listenWebSocket()

	case wsErrorMsg:
		c.liveError = msg.err.Error()
		c.liveConnected = false
		return c, nil

	case addCompanyMsg:
		if msg.err != nil {
			return c, func() tea.Msg {
				return messages.PageSwitchMsg{
					Err:  msg.err,
					Page: messages.ErrorPageNumber,
				}
			}
		} else {
			return c, func() tea.Msg {
				return messages.ReloadMsg{
					Page: messages.WatchlistPageNumber,
				}
			}
		}

	case tea.KeyMsg:
		if c.tabs[c.activeTab] == tabChart {
			return c.handleChartKeys(msg.String())
		}

		if c.tabs[c.activeTab] == tabLiveUpdate {
			return c.handleLiveChartKeys(msg.String())
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if c.ws != nil {
				c.ws.Close()
			}
			return c, tea.Quit

		case "esc":
			if c.ws != nil {
				c.ws.WriteMessage(websocket.TextMessage, []byte("exit"))
				c.ws.Close()
			}
			return c, func() tea.Msg {
				return messages.PageSwitchWithoutInitMsg{
					Page: c.PrevPage,
				}
			}

		case "right", "l":
			oldTab := c.activeTab
			if c.activeTab < len(c.tabs)-1 {
				c.activeTab++
			} else {
				c.activeTab = 0
			}

			// Initialize chart data when entering chart tab
			if c.tabs[c.activeTab] == tabChart && c.tabs[oldTab] != tabChart {
				c.chartLoading = true
				return c, c.fetchDataCmd()
			}

			// Initialize WebSocket when entering live tab
			if c.tabs[c.activeTab] == tabLiveUpdate && c.tabs[oldTab] != tabLiveUpdate {
				cmd := c.connectWebSocket()
				return c, tea.Batch(
					cmd,
					c.listenWebSocket(),
				)
			}

			// Clean up WebSocket when leaving live tab
			if c.tabs[oldTab] == tabLiveUpdate && c.tabs[c.activeTab] != tabLiveUpdate {
				if c.ws != nil {
					c.ws.WriteMessage(websocket.TextMessage, []byte("exit"))
					c.ws.Close()
					c.ws = nil
					c.liveConnected = false
				}
			}

		case "left", "h":
			oldTab := c.activeTab
			if c.activeTab > 0 {
				c.activeTab--
			} else {
				c.activeTab = len(c.tabs) - 1
			}

			// Initialize chart data when entering chart tab
			if c.tabs[c.activeTab] == tabChart && c.tabs[oldTab] != tabChart {
				c.chartLoading = true
				return c, c.fetchDataCmd()
			}

			// Initialize WebSocket when entering live tab
			if c.tabs[c.activeTab] == tabLiveUpdate && c.tabs[oldTab] != tabLiveUpdate {
				cmd := c.connectWebSocket()
				return c, tea.Batch(
					cmd,
					c.listenWebSocket(),
				)
			}

			// Clean up WebSocket when leaving live tab
			if c.tabs[oldTab] == tabLiveUpdate && c.tabs[c.activeTab] != tabLiveUpdate {
				if c.ws != nil {
					c.ws.WriteMessage(websocket.TextMessage, []byte("exit"))
					c.ws.Close()
					c.ws = nil
					c.liveConnected = false
				}
			}

		case "a", "A":
			return c, tea.Batch(c.addCompanyToWatchlist())

		case "b", "B":
			return c, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.BuyPageNumber,
				}
			}
		}
	}

	return c, nil
}

func (c CompanyPage) getDataFromBar(bar BarData) (string, string, string, string) {
	return fmt.Sprintf("%.2f", bar.Open), fmt.Sprintf("%.2f", bar.High), fmt.Sprintf("%.2f", bar.Low), fmt.Sprintf("%.2f", bar.Close)

}

func (c *CompanyPage) handleChartKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "1":
		c.timeFrame = TimeFrameMinute
		c.chartLoading = true
		return *c, c.fetchDataCmd()
	case "2":
		c.timeFrame = TimeFrameHour
		c.chartLoading = true
		return *c, c.fetchDataCmd()
	case "3":
		c.timeFrame = TimeFrameDay
		c.chartLoading = true
		return *c, c.fetchDataCmd()
	case "4":
		c.timeFrame = TimeFrameWeek
		c.chartLoading = true
		return *c, c.fetchDataCmd()
	case "5":
		c.timeFrame = TimeFrameMonth
		c.chartLoading = true
		return *c, c.fetchDataCmd()
	// case "ctrl+h", "ctrl+left":
	// 	step := max(1, c.zoomLevel/5)
	// 	if c.viewStart > 0 {
	// 		c.viewStart = max(0, c.viewStart-step)
	// 		c.viewEnd = c.viewStart + c.zoomLevel
	// 		if c.viewEnd > len(c.chartData) {
	// 			c.viewEnd = len(c.chartData)
	// 		}
	// 		c.redrawChart()
	// 	}
	// case "ctrl+l", "ctrl+right":
	// 	step := max(1, c.zoomLevel/5)
	// 	if c.viewEnd < len(c.chartData) {
	// 		c.viewEnd = min(len(c.chartData), c.viewEnd+step)
	// 		c.viewStart = c.viewEnd - c.zoomLevel
	// 		if c.viewStart < 0 {
	// 			c.viewStart = 0
	// 		}
	// 		c.redrawChart()
	// 	}
	// case "ctrl+k", "ctrl+up", "+":
	// 	if c.zoomLevel > 10 {
	// 		c.zoomLevel = max(10, c.zoomLevel-10)
	// 		c.viewStart += 10
	// 		c.viewEnd = min(len(c.chartData)-1, c.viewStart+c.zoomLevel)
	// 		c.redrawChart()
	// 	}
	// case "ctrl+j", "ctrl+down", "-":
	// 	maxZoom := len(c.chartData)
	// 	if c.zoomLevel < maxZoom {
	// 		c.zoomLevel = min(maxZoom, c.zoomLevel+10)
	// 		c.viewStart -= 10
	// 		c.viewEnd = min(len(c.chartData), c.viewStart+c.zoomLevel)
	// 		c.redrawChart()
	// 	}
	// case "home", "g":
	// 	c.viewStart = 0
	// 	c.viewEnd = min(len(c.chartData), c.zoomLevel)
	// 	c.redrawChart()
	// case "end", "G":
	// 	c.viewEnd = len(c.chartData)
	// 	c.viewStart = max(0, c.viewEnd-c.zoomLevel)
	// 	c.redrawChart()
	case "right", "l":
		oldTab := c.activeTab
		if c.activeTab < len(c.tabs)-1 {
			c.activeTab++
		} else {
			c.activeTab = 0
		}

		// Initialize chart data when entering chart tab
		if c.tabs[c.activeTab] == tabChart && c.tabs[oldTab] != tabChart {
			c.chartLoading = true
			return *c, c.fetchDataCmd()
		}

		// Initialize WebSocket when entering live tab
		if c.tabs[c.activeTab] == tabLiveUpdate && c.tabs[oldTab] != tabLiveUpdate {
			cmd := c.connectWebSocket()
			return *c, tea.Batch(
				cmd,
				c.listenWebSocket(),
			)
		}

		// Clean up WebSocket when leaving live tab
		if c.tabs[oldTab] == tabLiveUpdate && c.tabs[c.activeTab] != tabLiveUpdate {
			if c.ws != nil {
				c.ws.WriteMessage(websocket.TextMessage, []byte("exit"))
				c.ws.Close()
				c.ws = nil
				c.liveConnected = false
			}
		}

	case "left", "h":
		oldTab := c.activeTab
		if c.activeTab > 0 {
			c.activeTab--
		} else {
			c.activeTab = len(c.tabs) - 1
		}

		// Initialize chart data when entering chart tab
		if c.tabs[c.activeTab] == tabChart && c.tabs[oldTab] != tabChart {
			c.chartLoading = true
			return *c, c.fetchDataCmd()
		}

		// Initialize WebSocket when entering live tab
		if c.tabs[c.activeTab] == tabLiveUpdate && c.tabs[oldTab] != tabLiveUpdate {
			cmd := c.connectWebSocket()
			return *c, tea.Batch(
				cmd,
				c.listenWebSocket(),
			)
		}

		// Clean up WebSocket when leaving live tab
		if c.tabs[oldTab] == tabLiveUpdate && c.tabs[c.activeTab] != tabLiveUpdate {
			if c.ws != nil {
				c.ws.WriteMessage(websocket.TextMessage, []byte("exit"))
				c.ws.Close()
				c.ws = nil
				c.liveConnected = false
			}
		}
	case "a", "A":
		return c, c.addCompanyToWatchlist()

	case "b", "B":
		return c, func() tea.Msg {
			return messages.PageSwitchMsg{
				Page: messages.BuyPageNumber,
			}
		}

	case "q", "ctrl+c":
		if c.ws != nil {
			c.ws.Close()
		}
		return *c, tea.Quit

	case "esc":
		if c.ws != nil {
			c.ws.WriteMessage(websocket.TextMessage, []byte("exit"))
			c.ws.Close()
		}
		return *c, func() tea.Msg {
			return messages.PageSwitchWithoutInitMsg{
				Page: c.PrevPage,
			}
		}
	case "r":
		c.chartLoading = true
		return *c, c.fetchDataCmd()
	}
	return *c, nil
}

func (c *CompanyPage) handleLiveChartKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "r":
		if c.ws != nil {
			c.ws.Close()
		}
		c.liveData = make([]timeserieslinechart.TimePoint, 0)
		c.liveConnected = false
		c.liveError = ""
		cmd := c.connectWebSocket()
		return *c, tea.Batch(
			cmd,
			c.listenWebSocket(),
		)
	case "right", "l":
		oldTab := c.activeTab
		if c.activeTab < len(c.tabs)-1 {
			c.activeTab++
		} else {
			c.activeTab = 0
		}

		// Initialize chart data when entering chart tab
		if c.tabs[c.activeTab] == tabChart && c.tabs[oldTab] != tabChart {
			c.chartLoading = true
			return *c, c.fetchDataCmd()
		}

		// Initialize WebSocket when entering live tab
		if c.tabs[c.activeTab] == tabLiveUpdate && c.tabs[oldTab] != tabLiveUpdate {
			cmd := c.connectWebSocket()
			return *c, tea.Batch(
				cmd,
				c.listenWebSocket(),
			)
		}

		// Clean up WebSocket when leaving live tab
		if c.tabs[oldTab] == tabLiveUpdate && c.tabs[c.activeTab] != tabLiveUpdate {
			if c.ws != nil {
				c.ws.WriteMessage(websocket.TextMessage, []byte("exit"))
				c.ws.Close()
				c.ws = nil
				c.liveConnected = false
			}
		}

	case "left", "h":
		oldTab := c.activeTab
		if c.activeTab > 0 {
			c.activeTab--
		} else {
			c.activeTab = len(c.tabs) - 1
		}

		// Initialize chart data when entering chart tab
		if c.tabs[c.activeTab] == tabChart && c.tabs[oldTab] != tabChart {
			c.chartLoading = true
			return *c, c.fetchDataCmd()
		}

		// Initialize WebSocket when entering live tab
		if c.tabs[c.activeTab] == tabLiveUpdate && c.tabs[oldTab] != tabLiveUpdate {
			cmd := c.connectWebSocket()
			return *c, tea.Batch(
				cmd,
				c.listenWebSocket(),
			)
		}

		// Clean up WebSocket when leaving live tab
		if c.tabs[oldTab] == tabLiveUpdate && c.tabs[c.activeTab] != tabLiveUpdate {
			if c.ws != nil {
				c.ws.WriteMessage(websocket.TextMessage, []byte("exit"))
				c.ws.Close()
				c.ws = nil
				c.liveConnected = false
			}
		}
	case "a", "A":
		return c, c.addCompanyToWatchlist()

	case "b", "B":
		log.Println("Sending symbol: ", c.CompanyInfo.Symbol)
		return c, func() tea.Msg {
			return messages.PageSwitchMsg{
				Page:   messages.BuyPageNumber,
				Symbol: c.CompanyInfo.Symbol,
			}
		}

	case "q", "ctrl+c":
		if c.ws != nil {
			c.ws.Close()
		}
		return *c, tea.Quit

	case "esc":
		if c.ws != nil {
			c.ws.WriteMessage(websocket.TextMessage, []byte("exit"))
			c.ws.Close()
		}
		return *c, func() tea.Msg {
			return messages.PageSwitchWithoutInitMsg{
				Page: c.PrevPage,
			}
		}
	}

	return *c, nil
}

func (c CompanyPage) View() string {
	// Define colors
	cyan := lipgloss.Color("#00FFFF")
	purple := lipgloss.Color("#A020F0")
	gray := lipgloss.Color("#888888")

	// Tab styles
	activeTabStyle := lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true).
		Underline(true).
		Padding(0, 2)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(gray).
		Padding(0, 2)

	// Render tabs
	var tabsStr strings.Builder
	for i, t := range c.tabs {
		title := tabTitles[t]
		if i == c.activeTab {
			tabsStr.WriteString(activeTabStyle.Render(title))
		} else {
			tabsStr.WriteString(inactiveTabStyle.Render(title))
		}
		if i < len(c.tabs)-1 {
			tabsStr.WriteString("  ")
		}
	}

	// Content box style
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(2, 4).
		Width(c.BaseModel.Width - 10).
		Height(c.BaseModel.Height - 10)

	// Get content for active tab
	var content string
	switch c.tabs[c.activeTab] {
	case tabOverview:
		content = c.renderOverview()
	case tabHistory:
		content = c.renderHistory()
	case tabPrice:
		content = c.renderPrice()
	case tabChart:
		content = c.renderChart()
	case tabLiveUpdate:
		content = c.renderLivePrice()
	case tabWarning:
		content = c.renderWarning()
	}

	// Wrap content in styled box
	contentBox := contentStyle.Render(content)

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true).
		Padding(1, 0).
		Width(c.BaseModel.Width).
		Align(lipgloss.Center)

	title := titleStyle.Render(fmt.Sprintf("📊 %s (%s)", c.CompanyInfo.Name, c.CompanyInfo.Symbol))

	// Help text
	help := ""
	if c.tabs[c.activeTab] == tabChart {
		help = "[1-5] timeframe  [Ctrl+←/→] pan left/right  [Ctrl+↑/↓] pan up/down  [+/-] zoom  [r] refresh  [←/→] tabs  [a] add company to watchlist  [b] buy  [esc] back  [q] quit"
	} else if c.tabs[c.activeTab] == tabLiveUpdate {
		help = "[r] reconnect  [←/→] tabs  [a] add company to watchlist  [b] buy  [esc] back  [q] quit"
	} else {
		help = "← → / h l: switch tabs • a: add company to watchlist • b: buy • esc: back • q: quit"
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		Italic(true).
		Width(c.BaseModel.Width).
		Align(lipgloss.Center).
		MarginTop(1)

	helpText := helpStyle.Render(help)

	// Combine everything
	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		tabsStr.String(),
		contentBox,
		helpText,
	)
}

func (c CompanyPage) renderOverview() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00AAFF")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	var lines []string
	lines = append(lines, labelStyle.Render("Company Name:    ")+valueStyle.Render(c.CompanyInfo.Name))
	lines = append(lines, labelStyle.Render("Stock Symbol:    ")+valueStyle.Render(c.CompanyInfo.Symbol))
	lines = append(lines, labelStyle.Render("Domain:          ")+valueStyle.Render(c.CompanyInfo.Domain))
	lines = append(lines, labelStyle.Render("Founded:         ")+valueStyle.Render(fmt.Sprintf("%d", c.CompanyInfo.FoundedYear)))
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Description:"))
	lines = append(lines, valueStyle.Render(c.CompanyInfo.Description))

	return strings.Join(lines, "\n")
}

func (c CompanyPage) renderHistory() string {
	if c.CompanyInfo.History == "" {
		return "No history information available."
	}
	return c.CompanyInfo.History
}

func (c CompanyPage) renderPrice() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00AAFF")).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	var lines []string

	if c.CompanyInfo.OpeningPrice > 0 {
		lines = append(lines, labelStyle.Render("Opening Price:   ")+valueStyle.Render(fmt.Sprintf("$%.2f", c.CompanyInfo.OpeningPrice)))
	} else {
		lines = append(lines, labelStyle.Render("Opening Price:   ")+valueStyle.Render("N/A"))
	}

	if c.CompanyInfo.ClosingPrice > 0 {
		lines = append(lines, labelStyle.Render("Closing Price:   ")+valueStyle.Render(fmt.Sprintf("$%.2f", c.CompanyInfo.ClosingPrice)))
	} else {
		lines = append(lines, labelStyle.Render("Closing Price:   ")+valueStyle.Render("N/A"))
	}

	if c.CompanyInfo.OpeningPrice > 0 && c.CompanyInfo.ClosingPrice > 0 {
		change := c.CompanyInfo.ClosingPrice - c.CompanyInfo.OpeningPrice
		changePercent := (change / c.CompanyInfo.OpeningPrice) * 100

		changeStyle := lipgloss.NewStyle()
		if change > 0 {
			changeStyle = changeStyle.Foreground(lipgloss.Color("#00FF00"))
		} else if change < 0 {
			changeStyle = changeStyle.Foreground(lipgloss.Color("#FF0000"))
		} else {
			changeStyle = changeStyle.Foreground(lipgloss.Color("#FFFFFF"))
		}

		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Change:          ")+changeStyle.Render(fmt.Sprintf("$%.2f (%.2f%%)", change, changePercent)))
	}

	return strings.Join(lines, "\n")
}

func (c CompanyPage) renderChart() string {
	if c.chartLoading {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Bold(true).
			Render("Loading chart data...")
	}

	if c.chartError != "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Render("Error: " + c.chartError + "\n\nPress 'r' to retry")
	}

	if len(c.chartData) == 0 {
		return "No chart data available"
	}

	// Render timeframe selector
	var tfParts []string
	for i, label := range timeFrameLabels {
		if TimeFrame(i) == c.timeFrame {
			tfParts = append(tfParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true).
				Render(fmt.Sprintf("[%d] %s", i+1, label)))
		} else {
			tfParts = append(tfParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Render(fmt.Sprintf("[%d] %s", i+1, label)))
		}
	}
	timeFrameSelector := strings.Join(tfParts, "  ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		timeFrameSelector,
		"",
		c.chart.View(),
		"",
	)
}

func (c CompanyPage) renderLivePrice() string {
	if !c.liveConnected && c.liveError == "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Bold(true).
			Render("Connecting to live market data...")
	}

	if c.liveError != "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Render("Error: " + c.liveError + "\n\nPress 'r' to reconnect")
	}

	// Live indicator
	indicator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true).
		Render("🔴 LIVE")

	// Stats
	stats := ""
	if len(c.liveData) > 0 {
		changeColor := lipgloss.Color("#00FF00")
		changeSymbol := "▲"
		if c.lastChange < 0 {
			changeColor = lipgloss.Color("#FF0000")
			changeSymbol = "▼"
		}

		changePercent := 0.0
		if c.liveData[0].Value > 0 {
			changePercent = (c.lastChange / c.liveData[0].Value) * 100
		}

		stats = lipgloss.NewStyle().Foreground(changeColor).Render(
			fmt.Sprintf("Last: $%.2f  |  %s $%.2f (%.2f%%)  |  High: $%.2f  |  Low: $%.2f  |  Volume: %.0f  |  Points: %d",
				c.lastPrice, changeSymbol, c.lastChange, changePercent,
				c.high, c.low, c.volume, len(c.liveData)))
	} else {
		stats = "Waiting for market data..."
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		indicator,
		"",
		c.liveChart.View(),
		"",
		stats,
	)
}

func (c CompanyPage) renderWarning() string {
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true).
		Align(lipgloss.Center)

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA00"))

	return lipgloss.JoinVertical(
		lipgloss.Center,
		warningStyle.Render("⚠️  WARNING  ⚠️"),
		"",
		textStyle.Render("This company contains NSFW content."),
		textStyle.Render("Viewer discretion is advised."),
	)
}

func (c *CompanyPage) processWebSocketData(msg WebSocketMsg) error {
	t, err := time.Parse(time.RFC3339, msg.Time)
	if err != nil {
		return err
	}

	point := timeserieslinechart.TimePoint{
		Time:  t,
		Value: msg.Price,
	}
	c.liveData = append(c.liveData, point)

	// Keep only last 200 points
	if len(c.liveData) > 200 {
		c.liveData = c.liveData[len(c.liveData)-200:]
	}

	c.lastPrice = msg.Price
	if len(c.liveData) > 1 {
		c.lastChange = c.lastPrice - c.liveData[0].Value
	}

	if c.high == 0 || msg.Price > c.high {
		c.high = msg.Price
	}
	if c.low == 0 || msg.Price < c.low {
		c.low = msg.Price
	}

	c.volume += msg.Size

	c.liveChart = timeserieslinechart.New(120, 30, timeserieslinechart.WithStyle(
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	))

	for _, point := range c.liveData {
		c.liveChart.Push(point)
	}

	c.liveChart.Draw()

	return nil
}

func (c *CompanyPage) fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		start := ""
		switch c.timeFrame {
		case TimeFrameMinute:
			start = time.Now().UTC().Add(-time.Hour * 24).Format(time.RFC3339) // 1 day
			log.Println("Minute")
		case TimeFrameHour:
			log.Println("Hour")
			start = time.Now().UTC().Truncate(time.Hour * 24).Add(-time.Hour * 24 * 14).Format(time.RFC3339) // 14 days
		case TimeFrameDay:
			log.Println("Day")
			start = time.Now().UTC().Truncate(time.Hour * 24).Add(-time.Hour * 24 * 365).Format(time.RFC3339) // 1 year
		case TimeFrameWeek:
			log.Println("Week")
			start = time.Now().UTC().Truncate(time.Hour * 24).Add(-time.Hour * 24 * 365 * 6).Format(time.RFC3339) // 6 years
		case TimeFrameMonth:
			log.Println("Month")
			start = time.Now().UTC().Truncate(time.Hour * 24).Add(-time.Hour * 24 * 365 * 20).Format(time.RFC3339) // 20 years
		}

		url := fmt.Sprintf(
			"%s/data/bars?symbols=%s&start=%s&timeframe=%s",
			requests.BaseURL,
			c.CompanyInfo.Symbol,
			start,
			timeFrameStrings[c.timeFrame],
		)

		body, err := requests.MakeRequest(
			http.MethodGet,
			url,
			nil,
			http.DefaultClient,
			c.BaseModel.Token,
		)
		if err != nil {
			return fetchDataMsg{err: err}
		}

		var response BarsResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return fetchDataMsg{err: err}
		}

		data, ok := response.Bars[c.CompanyInfo.Symbol]
		if !ok || len(data) == 0 {
			return fetchDataMsg{err: fmt.Errorf("no data available for symbol %s", c.CompanyInfo.Symbol)}
		}

		return fetchDataMsg{data: data}
	}
}

func (c *CompanyPage) connectWebSocket() tea.Cmd {
	url := fmt.Sprintf("ws://localhost:42069/data/stocks/live/%s", c.CompanyInfo.Symbol)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return func() tea.Msg {
			return wsErrorMsg{err: fmt.Errorf("failed to connect: %w", err)}
		}
	}

	c.ws = ws
	return func() tea.Msg {
		return wsConnectedMsg{}
	}
}

func (c *CompanyPage) listenWebSocket() tea.Cmd {
	return func() tea.Msg {
		if c.ws == nil {
			return wsErrorMsg{err: fmt.Errorf("websocket not connected")}
		}

		var msg map[string]interface{}
		err := c.ws.ReadJSON(&msg)
		if err != nil {
			return wsErrorMsg{err: fmt.Errorf("connection lost: %w", err)}
		}

		if errMsg, ok := msg["error"].(string); ok {
			return wsErrorMsg{err: fmt.Errorf(errMsg)}
		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			return wsErrorMsg{err: err}
		}

		var wsMsg WebSocketMsg
		if err := json.Unmarshal(msgBytes, &wsMsg); err != nil {
			return wsErrorMsg{err: err}
		}

		return wsDataMsg{data: wsMsg}
	}
}

func (c CompanyPage) addCompanyToWatchlist() tea.Cmd {
	return func() tea.Msg {
		_, err := requests.MakeRequest(http.MethodPost, requests.BaseURL+"/watchlist/"+c.CompanyInfo.Symbol, nil, http.DefaultClient, c.BaseModel.Token)
		if err != nil {
			return addCompanyMsg{err: err}
		}

		return addCompanyMsg{err: nil}
	}
}

func (c *CompanyPage) Reload() {
	if c.ws != nil {
		c.ws.Close()
	}

	c.CompanyInfo = nil
	c.chartData = nil
	// c.zoomLevel = 100
	c.liveData = make([]timeserieslinechart.TimePoint, 0)
	c.chartLoading = false
	c.liveConnected = false
	c.chartError = ""
	c.liveError = ""
}
