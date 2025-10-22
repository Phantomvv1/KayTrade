package watchlistpage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WatchlistPage struct {
	BaseModel basemodel.BaseModel
	titleBar  string
	companies list.Model
	movers    []MarketMover
	cursor    int
	help      help.Model
	loadErr   error
	loaded    bool
	spinner   spinner.Model
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
	s := spinner.New()
	s.Spinner = spinner.Line
	spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	s.Style = spinnerStyle

	return WatchlistPage{
		BaseModel: basemodel.BaseModel{Client: client},
		help:      help.New(),
		cursor:    0,
		titleBar:  "📈 Watchlist",
		companies: l,
		spinner:   s,
	}
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Help   key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select the company"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (w WatchlistPage) init() tea.Msg {
	log.Println("Init works")
	wg := sync.WaitGroup{}
	wg.Add(2)

	var companies []CompanyInfo
	var movers MarketMovers
	var err1, err2 error

	go func() {
		defer wg.Done()
		resp, err := http.DefaultClient.Get("http://localhost:42069/watchlist/info")
		if err != nil {
			err1 = err
			return
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		json.Unmarshal(data, &companies)
		log.Println("Done from user's watchlist")
	}()

	go func() {
		defer wg.Done()
		resp, err := http.DefaultClient.Get("http://localhost:42069/data/stocks/top-market-movers?top=5")
		if err != nil {
			err2 = err
			return
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		json.Unmarshal(data, &movers)
		log.Println("Done from top market movers")
	}()

	go func() {
		wg.Wait()
	}()

	if err1 != nil || err2 != nil {
		return errors.New("Error unable to fetch the data")
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
		w.loadErr = msg
	case initResult:
		w.loaded = true
		var items []list.Item
		for _, company := range msg.Companies {
			items = append(items, companyItem{company: company})
		}

		w.movers = append(msg.Gainers, msg.Losers...)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return w, tea.Quit
		case key.Matches(msg, keys.Up):
			if w.cursor > 0 {
				w.cursor--
			}
		case key.Matches(msg, keys.Down):
			if w.cursor < len(w.companies.Items())-1 {
				w.cursor++
			}
		case key.Matches(msg, keys.Select):
			return w, tea.Quit
		case key.Matches(msg, keys.Help):
			w.help.ShowAll = !w.help.ShowAll
		}
	case spinner.TickMsg:
		log.Println("Spinner message")
		var cmd tea.Cmd
		w.spinner, cmd = w.spinner.Update(msg)
		return w, cmd
	}

	return w, nil
}
func (w WatchlistPage) View() string {
	cyan := lipgloss.Color("#00FFFF")
	purple := lipgloss.Color("#A020F0")
	errRed := lipgloss.Color("#ED2923")
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

	if w.loadErr != nil {
		return lipgloss.PlaceHorizontal(w.BaseModel.Width, lipgloss.Center,
			lipgloss.NewStyle().Foreground(errRed).Render("Error loading data: "+w.loadErr.Error()))
	}

	if !w.loaded {
		return lipgloss.Place(w.BaseModel.Width, w.BaseModel.Height, lipgloss.Center, lipgloss.Center, w.spinner.View())
	}

	// Right panel (Top movers)
	moverCards := []string{"Top Market Movers:\n"}
	for i, m := range w.movers {
		changeColor := green
		if i >= 5 {
			changeColor = red
		}

		line := fmt.Sprintf("%s  ", m.Symbol)
		price := lipgloss.NewStyle().Foreground(changeColor).Render(fmt.Sprintf("%.4f ", m.Price))
		line += price + fmt.Sprintf("%.2f %.2f", m.Change, m.PercentChange)
		moverCards = append(moverCards, line)
	}

	right := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(1, 2).
		Render(strings.Join(moverCards, "\n"))

	content := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(w.BaseModel.Width/2-2).Render(w.companies.View()),
		lipgloss.NewStyle().Width(w.BaseModel.Width/2-2).Render(right),
	)

	helpView := w.help.View(keys)
	height := 8 - strings.Count(helpView, "\n")

	header = lipgloss.PlaceHorizontal(w.BaseModel.Width, lipgloss.Center, header)

	return header + content + strings.Repeat("\n", height) + helpView
}
