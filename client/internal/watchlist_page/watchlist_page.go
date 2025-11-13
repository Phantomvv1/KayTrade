package watchlistpage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WatchlistPage struct {
	BaseModel      basemodel.BaseModel
	titleBar       string
	companies      list.Model
	movers         []MarketMover
	loaded         bool
	spinner        spinner.Model
	emptyWatchlist bool
	focusedList    bool
	renderedLogo   string
}

type CompanyInfo struct {
	Symbol       string  `json:"symbol"`
	OpeningPrice float64 `json:"opening_price,omitempty"`
	ClosingPrice float64 `json:"closing_price,omitempty"`
	Logo         string  `json:"logo"`
	Name         string  `json:"name"`
	History      string  `json:"history"`
	IsNSFW       bool    `json:"isNsfw"`
	Description  string  `json:"description"`
	FoundedYear  int     `json:"founded_year"`
	Domain       string  `json:"domain"`
}

func (c CompanyInfo) SymbolInfo() string {
	return c.Symbol
}

func (c CompanyInfo) OpeningPriceInfo() float64 {
	return c.OpeningPrice
}

func (c CompanyInfo) ClosingPriceInfo() float64 {
	return c.ClosingPrice
}

func (c CompanyInfo) LogoInfo() string {
	return c.Logo
}

func (c CompanyInfo) NameInfo() string {
	return c.Name
}

func (c CompanyInfo) HistoryInfo() string {
	return c.History
}

func (c CompanyInfo) IsNSFWInfo() bool {
	return c.IsNSFW
}

func (c CompanyInfo) DescriptionInfo() string {
	return c.Description
}

func (c CompanyInfo) FoundedYearInfo() int {
	return c.FoundedYear
}

func (c CompanyInfo) DomainInfo() string {
	return c.Domain
}

type MarketMover struct {
	Change        float64 `json:"change"`
	PercentChange float64 `json:"percent_change"`
	Price         float64 `json:"price"`
	Symbol        string  `json:"symbol"`
}

type MarketMovers struct {
	Gainers []MarketMover `json:"gainers"`
	Losers  []MarketMover `json:"losers"`
	Updated string        `json:"last_updated"`
}

type initResult struct {
	Companies []CompanyInfo
	Gainers   []MarketMover
	Losers    []MarketMover
	Updated   time.Time
}

type companyItem struct {
	company CompanyInfo
}

func (c companyItem) Title() string       { return c.company.Name }
func (c companyItem) Description() string { return c.company.Description }
func (c companyItem) FilterValue() string { return c.company.Name }

func NewWatchlistPage(client *http.Client) WatchlistPage {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.KeyMap.Quit = key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit"))

	s := spinner.New()
	s.Spinner = spinner.Line
	spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	s.Style = spinnerStyle

	return WatchlistPage{
		BaseModel:      basemodel.BaseModel{Client: client},
		titleBar:       "📈 Watchlist",
		companies:      l,
		spinner:        s,
		emptyWatchlist: false,
		focusedList:    true,
	}
}

func (w WatchlistPage) init() tea.Msg {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var companies []CompanyInfo
	var movers MarketMovers
	var err1, err2 error

	go func() {
		defer wg.Done()
		body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/watchlist/info", nil, http.DefaultClient, w.BaseModel.Token)
		if err != nil {
			err1 = err
			return
		}

		var info map[string][]CompanyInfo
		err1 = json.Unmarshal(body, &info)
		companies = info["information"]
	}()

	go func() {
		defer wg.Done()
		body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/data/stocks/top-market-movers?top=5", nil, http.DefaultClient, "")
		if err != nil {
			err2 = err
			return
		}
		err2 = json.Unmarshal(body, &movers)
	}()

	wg.Wait()

	if err1 != nil || err2 != nil {
		if err1 != nil {
			return err1
		} else {
			return err2
		}
	}

	updated, err := time.Parse(time.RFC3339, movers.Updated)
	if err != nil {
		return err
	}

	return initResult{Companies: companies, Gainers: movers.Gainers, Losers: movers.Losers, Updated: updated}
}

func (w WatchlistPage) Init() tea.Cmd {
	return tea.Batch(w.init, w.spinner.Tick)
}

func (w WatchlistPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case error:
		return w, func() tea.Msg {
			return messages.PageSwitchMsg{
				Page: messages.ErrorPageNumber,
				Err:  msg,
			}
		}
	case initResult:
		w.loaded = true
		for i, company := range msg.Companies {
			w.companies.InsertItem(i, companyItem{company: company})
		}
		w.companies.SetSize(w.BaseModel.Width/2-4, w.BaseModel.Height-8)

		if len(w.companies.Items()) == 0 {
			w.emptyWatchlist = true
		}

		w.movers = append(msg.Gainers, msg.Losers...)
		w.companies.FilterInput.Focus()
	case tea.KeyMsg:
		var cmd tea.Cmd
		w.companies, cmd = w.companies.Update(msg)
		if msg.String() == "s" {
			return w, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.SearchPage,
				}
			}
		}

		return w, cmd
	case spinner.TickMsg:
		var cmd tea.Cmd
		w.spinner, cmd = w.spinner.Update(msg)
		return w, cmd
	}

	return w, nil
}
func (w WatchlistPage) View() string {
	cyan := lipgloss.Color("#00FFFF")
	purple := lipgloss.Color("#A020F0")
	red := lipgloss.Color("#D30000")
	green := lipgloss.Color("#0B6623")

	headerStyle := lipgloss.NewStyle().
		Foreground(purple).
		Background(cyan).
		Bold(true).
		Padding(0, 2).
		MarginBottom(1).
		Align(lipgloss.Center)

	header := headerStyle.Render(w.titleBar) + "\n\n"

	if !w.loaded {
		return lipgloss.Place(w.BaseModel.Width, w.BaseModel.Height, lipgloss.Center, lipgloss.Center, w.spinner.View())
	}

	// Right panel (Top movers)
	moverCards := []string{"Top Market Movers:\n"}
	moverCards = append(moverCards, "ticker price change %change")
	for i, m := range w.movers {
		changeColor := green
		if i >= 5 {
			changeColor = red
		}

		line := fmt.Sprintf("%s"+strings.Repeat(" ", 7-len(m.Symbol)), m.Symbol) // allign price
		p := fmt.Sprintf("%.2f", m.Price)
		price := lipgloss.NewStyle().Foreground(changeColor).Render(p + strings.Repeat(" ", 6-len(p))) // allign change
		c := fmt.Sprintf("%.2f", m.Change)
		line += price + fmt.Sprintf(c+strings.Repeat(" ", 7-len(c))+"%.2f", m.PercentChange) // allign percent change
		moverCards = append(moverCards, line)
	}

	right := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(1, 2).
		Render(strings.Join(moverCards, "\n"))

	content := ""
	if !w.emptyWatchlist {
		content = lipgloss.JoinHorizontal(lipgloss.Top,
			// lipgloss.NewStyle().MarginLeft(1).Render(w.renderedLogo),
			lipgloss.NewStyle().Width(w.BaseModel.Width/2-2).MarginLeft(1).Render(w.companies.View()),
			lipgloss.NewStyle().Width(w.BaseModel.Width/2-2).MarginLeft(20).Render(right),
		)

	} else {
		msg := lipgloss.NewStyle().
			Padding(1, 1).
			MarginLeft(20).
			Render("There are no companies in your watchlist.\nAdd some in order to see them here.")

		content = lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().Width(w.BaseModel.Width/2-2).Render(msg),
			lipgloss.NewStyle().Width(w.BaseModel.Width/2-2).Render(right),
		)
	}

	header = lipgloss.PlaceHorizontal(w.BaseModel.Width, lipgloss.Center, header)

	return header + content
}
