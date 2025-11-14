package companypage

import (
	"fmt"
	"net/http"
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	tabOverview = iota
	tabHistory
	tabPrice
	tabWarning
)

var tabTitles = []string{
	"Overview",
	"History",
	"Price",
	"Warning",
}

type CompanyPage struct {
	BaseModel   basemodel.BaseModel
	CompanyInfo *messages.CompanyInfo
	activeTab   int
	tabs        []int
}

func NewCompanyPage(client *http.Client) CompanyPage {
	return CompanyPage{
		BaseModel: basemodel.BaseModel{Client: client},
		activeTab: tabOverview,
		tabs:      []int{tabOverview, tabHistory, tabPrice},
	}
}

func (c CompanyPage) Init() tea.Cmd {
	return nil
}

func (c CompanyPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if c.CompanyInfo.IsNSFW && c.tabs[len(c.tabs)-1] != tabWarning {
		c.tabs = append(c.tabs, tabWarning)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return c, tea.Quit

		case "esc":
			return c, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.SearchPageNumber,
				}
			}

		case "right", "l":
			if c.activeTab < len(c.tabs)-1 {
				c.activeTab++
			} else {
				c.activeTab = 0
			}

		case "left", "h":
			if c.activeTab > 0 {
				c.activeTab--
			} else {
				c.activeTab = len(c.tabs) - 1
			}
		}
	}

	return c, nil
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
	helpStyle := lipgloss.NewStyle().
		Foreground(gray).
		Italic(true).
		Width(c.BaseModel.Width).
		Align(lipgloss.Center).
		MarginTop(1)

	help := helpStyle.Render("← → / h l: switch tabs  •  esc: back  •  q: quit")

	// Combine everything
	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		tabsStr.String(),
		contentBox,
		help,
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

func (c *CompanyPage) Reload() {
	c.activeTab = 0
	c.CompanyInfo = nil
}
