package companypage

import (
	"encoding/json"
	"fmt"
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

	// Chart state
	chart        timeserieslinechart.Model
	timeFrame    TimeFrame
	chartData    []BarData
	viewStart    int
	viewEnd      int
	zoomLevel    int
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

func NewCompanyPage(client *http.Client) CompanyPage {
	// Create historical chart
	chart := timeserieslinechart.New(80, 20,
		timeserieslinechart.WithStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		))

	// Create live chart
	liveChart := timeserieslinechart.New(80, 20,
		timeserieslinechart.WithStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		))

	return CompanyPage{
		BaseModel: basemodel.BaseModel{Client: client},
		activeTab: tabOverview,
		tabs:      []int{tabOverview, tabHistory, tabPrice, tabChart, tabLiveUpdate},
		chart:     chart,
		liveChart: liveChart,
		timeFrame: TimeFrameDay,
		zoomLevel: 100,
		liveData:  make([]timeserieslinechart.TimePoint, 0),
	}
}

func (c CompanyPage) Init() tea.Cmd {
	return nil
}

func (c CompanyPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if c.CompanyInfo != nil && c.CompanyInfo.IsNSFW && c.tabs[len(c.tabs)-1] != tabWarning {
		c.tabs = append(c.tabs, tabWarning)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.BaseModel.Width = msg.Width
		c.BaseModel.Height = msg.Height
		c.chart.Resize(msg.Width-10, msg.Height-20)
		c.liveChart.Resize(msg.Width-10, msg.Height-20)
		return c, nil

	case fetchDataMsg:
		c.chartLoading = false
		if msg.err != nil {
			c.chartError = msg.err.Error()
			return c, nil
		}
		c.chartData = msg.data
		if len(c.chartData) > 0 {
			c.viewEnd = len(c.chartData)
			c.viewStart = max(0, c.viewEnd-c.zoomLevel)
			c.updateChart()
		}
		return c, nil

	case wsConnectedMsg:
		c.liveConnected = true
		return c, nil

	case wsDataMsg:
		c.processWebSocketData(msg.data)
		c.updateLiveChart()
		return c, c.listenWebSocket()

	case wsErrorMsg:
		c.liveError = msg.err.Error()
		c.liveConnected = false
		return c, nil

	case tea.KeyMsg:
		// Handle chart-specific controls when on chart tab
		if c.tabs[c.activeTab] == tabChart {
			return c.handleChartKeys(msg.String())
		}

		// Handle live chart controls when on live tab
		if c.tabs[c.activeTab] == tabLiveUpdate {
			return c.handleLiveChartKeys(msg.String())
		}

		// Handle general navigation
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
				return c, tea.Batch(
					c.connectWebSocket(),
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
				return c, tea.Batch(
					c.connectWebSocket(),
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
		}
	}

	return c, nil
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
	case "ctrl+h", "ctrl+left":
		c.panLeft()
		c.updateChart()
	case "ctrl+l", "ctrl+right":
		c.panRight()
		c.updateChart()
	case "ctrl+k", "ctrl+up", "+":
		c.zoomIn()
		c.updateChart()
	case "ctrl+j", "ctrl+down", "-":
		c.zoomOut()
		c.updateChart()
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
			return c, tea.Batch(
				c.connectWebSocket(),
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
			return c, tea.Batch(
				c.connectWebSocket(),
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
	case "q", "ctrl+c":
		if c.ws != nil {
			c.ws.Close()
		}
		return c, tea.Quit
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
		return *c, tea.Batch(
			c.connectWebSocket(),
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
			return c, c.fetchDataCmd()
		}

		// Initialize WebSocket when entering live tab
		if c.tabs[c.activeTab] == tabLiveUpdate && c.tabs[oldTab] != tabLiveUpdate {
			return c, tea.Batch(
				c.connectWebSocket(),
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
			return c, tea.Batch(
				c.connectWebSocket(),
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
	case "q", "ctrl+c":
		if c.ws != nil {
			c.ws.Close()
		}
		return c, tea.Quit
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
	var help string
	if c.tabs[c.activeTab] == tabChart {
		help = "[1-5] timeframe  [Ctrl+←/→] pan  [Ctrl+↑/↓] zoom  [r] refresh  [←/→] tabs  [esc] back  [q] quit"
	} else if c.tabs[c.activeTab] == tabLiveUpdate {
		help = "[r] reconnect  [←/→] tabs  [esc] back  [q] quit"
	} else {
		help = "← → / h l: switch tabs  •  esc: back  •  q: quit"
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

	// Render stats
	visible := c.getVisibleData()
	stats := ""
	if len(visible) > 0 {
		latest := visible[len(visible)-1]
		first := visible[0]
		change := latest.Close - first.Open
		changePercent := (change / first.Open) * 100

		changeColor := lipgloss.Color("#00FF00")
		changeSymbol := "▲"
		if change < 0 {
			changeColor = lipgloss.Color("#FF0000")
			changeSymbol = "▼"
		}

		stats = lipgloss.NewStyle().Foreground(changeColor).Render(
			fmt.Sprintf("Price: $%.2f  |  %s %.2f (%.2f%%)  |  High: $%.2f  |  Low: $%.2f  |  Range: %d-%d of %d",
				latest.Close, changeSymbol, change, changePercent,
				c.findHigh(visible), c.findLow(visible),
				c.viewStart+1, c.viewEnd, len(c.chartData)))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		timeFrameSelector,
		"",
		c.chart.View(),
		"",
		stats,
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

// Chart helper methods
func (c *CompanyPage) updateChart() {
	if len(c.chartData) == 0 {
		return
	}

	visible := c.getVisibleData()
	if len(visible) == 0 {
		return
	}

	c.chart.Clear()

	for _, bar := range visible {
		c.chart.Push(timeserieslinechart.TimePoint{Time: bar.Timestamp, Value: bar.Close})
	}

}

func (c *CompanyPage) updateLiveChart() {
	if len(c.liveData) == 0 {
		return
	}

	c.liveChart.Clear()

	var points []timeserieslinechart.TimePoint
	for _, point := range c.liveData {
		points = append(points, point)
	}
}

func (c *CompanyPage) getVisibleData() []BarData {
	if c.viewStart >= len(c.chartData) {
		return nil
	}
	return c.chartData[c.viewStart:c.viewEnd]
}

func (c *CompanyPage) panLeft() {
	step := c.zoomLevel / 4
	if step < 1 {
		step = 1
	}
	c.viewStart = max(0, c.viewStart-step)
	c.viewEnd = min(len(c.chartData), c.viewStart+c.zoomLevel)
}

func (c *CompanyPage) panRight() {
	step := c.zoomLevel / 4
	if step < 1 {
		step = 1
	}
	c.viewEnd = min(len(c.chartData), c.viewEnd+step)
	c.viewStart = max(0, c.viewEnd-c.zoomLevel)
}

func (c *CompanyPage) zoomIn() {
	c.zoomLevel = max(10, c.zoomLevel-10)
	center := (c.viewStart + c.viewEnd) / 2
	c.viewStart = max(0, center-c.zoomLevel/2)
	c.viewEnd = min(len(c.chartData), c.viewStart+c.zoomLevel)
}

func (c *CompanyPage) zoomOut() {
	c.zoomLevel = min(len(c.chartData), c.zoomLevel+10)
	center := (c.viewStart + c.viewEnd) / 2
	c.viewStart = max(0, center-c.zoomLevel/2)
	c.viewEnd = min(len(c.chartData), c.viewStart+c.zoomLevel)
}

func (c *CompanyPage) findHigh(data []BarData) float64 {
	if len(data) == 0 {
		return 0
	}
	high := data[0].High
	for _, bar := range data {
		if bar.High > high {
			high = bar.High
		}
	}
	return high
}

func (c *CompanyPage) findLow(data []BarData) float64 {
	if len(data) == 0 {
		return 0
	}
	low := data[0].Low
	for _, bar := range data {
		if bar.Low < low {
			low = bar.Low
		}
	}
	return low
}

func (c *CompanyPage) processWebSocketData(msg WebSocketMsg) {
	now := time.Now()

	point := timeserieslinechart.TimePoint{
		Time:  now,
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
}

func (c *CompanyPage) fetchDataCmd() tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf(
			"%s/data/bars?symbols=%s&start=&timeframe=%s",
			requests.BaseURL,
			c.CompanyInfo.Symbol,
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
	return func() tea.Msg {
		url := fmt.Sprintf("ws://localhost:42069/data/real-time/%s", c.CompanyInfo.Symbol)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			return wsErrorMsg{err: fmt.Errorf("failed to connect: %w", err)}
		}

		c.ws = ws
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

func (c *CompanyPage) Reload() {
	if c.ws != nil {
		c.ws.Close()
	}
	c.activeTab = 0
	c.CompanyInfo = nil
	c.chartData = nil
	c.liveData = make([]timeserieslinechart.TimePoint, 0)
	c.chartLoading = false
	c.liveConnected = false
	c.chartError = ""
	c.liveError = ""
}
