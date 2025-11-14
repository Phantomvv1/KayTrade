package companypage

import (
	"fmt"
	"net/http"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	tabOverview = iota
	tabDescription
	tabHistory
	tabPrice
	tabWarning
)

var tabTitles = []string{
	"Overview",
	"Description",
	"History",
	"Price",
	"Warning",
}

type CompanyPage struct {
	BaseModel   basemodel.BaseModel
	CompanyInfo *messages.CompanyInfo
	activeTab   int
	viewport    viewport.Model
	tabs        []int
}

func NewCompanyPage(client *http.Client) CompanyPage {
	vp := viewport.New(100, 70)
	vp.Style = lipgloss.NewStyle().MarginTop(1)

	return CompanyPage{
		BaseModel: basemodel.BaseModel{Client: client},
		activeTab: tabOverview,
		viewport:  vp,
		tabs:      []int{tabOverview, tabDescription, tabHistory, tabPrice},
	}
}

func (c *CompanyPage) refreshViewport() {
	switch c.activeTab {
	case tabOverview:
		c.viewport.SetContent(c.renderOverview())

	case tabDescription:
		c.viewport.SetContent(c.CompanyInfo.Description)

	case tabHistory:
		c.viewport.SetContent(c.CompanyInfo.History)

	case tabPrice:
		c.viewport.SetContent(c.renderPrice())

	case tabWarning:
		c.viewport.SetContent(c.renderWarning())
	}
}

func (c CompanyPage) Init() tea.Cmd {
	if c.CompanyInfo != nil {
		c.refreshViewport()
	}

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
				return messages.PageSwitchWithoutInitMsg{
					Page: messages.WatchlistPageNumber,
				}
			}

		case "right", "l":
			c.nextTab()
		case "left", "h":
			c.previousTab()

		case "down", "j", "up", "k":
			var cmd tea.Cmd
			c.viewport, cmd = c.viewport.Update(msg)
			return c, cmd
		}
	}

	return c, nil
}

func (c *CompanyPage) nextTab() {
	if c.activeTab < len(c.tabs)-1 {
		c.activeTab++
	} else {
		c.activeTab = 0
	}
	c.refreshViewport()
}

func (c *CompanyPage) previousTab() {
	if c.activeTab > 0 {
		c.activeTab--
	} else {
		c.activeTab = len(c.tabs) - 1
	}
	c.refreshViewport()
}

func (c CompanyPage) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		c.renderTabs(),
		c.viewport.View(),
	)
}

func (c CompanyPage) renderTabs() string {
	var out string

	activeStyle := lipgloss.NewStyle().
		MarginRight(2).
		Foreground(lipgloss.Color("#00AAFF")).
		Underline(true)

	inactiveStyle := lipgloss.NewStyle().
		MarginRight(2).
		Foreground(lipgloss.Color("#888888"))

	for _, t := range c.tabs {
		title := tabTitles[t]
		if t == c.activeTab {
			out += activeStyle.Render(title)
		} else {
			out += inactiveStyle.Render(title)
		}
	}

	return out
}

func (c CompanyPage) renderOverview() string {
	lines := ""
	lines += fmt.Sprintf("Name: %s\n", c.CompanyInfo.Name)
	lines += fmt.Sprintf("Symbol: %s\n", c.CompanyInfo.Symbol)
	lines += fmt.Sprintf("Domain: %s\n", c.CompanyInfo.Domain)
	lines += fmt.Sprintf("Founded: %d\n", c.CompanyInfo.FoundedYear)
	lines += fmt.Sprintf("Logo: %s\n", c.CompanyInfo.Logo)

	return lines
}

func (c CompanyPage) renderPrice() string {
	out := ""
	out += fmt.Sprintf("Opening Price: %.2f\n", c.CompanyInfo.OpeningPrice)
	out += fmt.Sprintf("Closing Price: %.2f\n", c.CompanyInfo.ClosingPrice)
	return out
}

func (c CompanyPage) renderWarning() string {
	if !c.CompanyInfo.IsNSFW {
		return ""
	}
	return "⚠ WARNING: This company contains NSFW content."
}

func (c *CompanyPage) Reload() {
	c.activeTab = 0
	c.CompanyInfo = nil
}
