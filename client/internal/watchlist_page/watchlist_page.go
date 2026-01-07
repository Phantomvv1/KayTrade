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
	filtering      bool
	Reloaded       bool
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
	Companies []messages.CompanyInfo
	Gainers   []MarketMover
	Losers    []MarketMover
	Updated   time.Time
}

type companyItem struct {
	company messages.CompanyInfo
}

func (c companyItem) Title() string       { return c.company.Name }
func (c companyItem) Description() string { return c.company.Description }
func (c companyItem) FilterValue() string { return c.company.Name }

func NewWatchlistPage(client *http.Client, tokenStore *basemodel.TokenStore) WatchlistPage {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("s", "S"), key.WithHelp("s", "search")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
			key.NewBinding(key.WithKeys("r", "R"), key.WithHelp("r", "remove company")),
			key.NewBinding(key.WithKeys("d", "D"), key.WithHelp("d", "remove all companies")),
			key.NewBinding(key.WithKeys("p", "P"), key.WithHelp("p", "profile page")),
			key.NewBinding(key.WithKeys("b", "B"), key.WithHelp("b", "bank relationship page")),
		}
	}

	s := spinner.New()
	s.Spinner = spinner.Line
	spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	s.Style = spinnerStyle

	return WatchlistPage{
		BaseModel:      basemodel.BaseModel{Client: client, TokenStore: tokenStore},
		titleBar:       "📈 Watchlist",
		companies:      l,
		spinner:        s,
		emptyWatchlist: false,
		focusedList:    true,
		filtering:      false,
		Reloaded:       false,
	}
}

func (w WatchlistPage) init() tea.Msg {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var companies []messages.CompanyInfo
	var movers MarketMovers
	var err1, err2 error

	go func() {
		defer wg.Done()
		body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/watchlist/info", nil, w.BaseModel.Client, w.BaseModel.TokenStore)
		if err != nil {
			err1 = err
			return
		}

		var info map[string][]messages.CompanyInfo
		err1 = json.Unmarshal(body, &info)
		companies = info["information"]
	}()

	go func() {
		defer wg.Done()
		body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/data/stocks/top-market-movers?top=5", nil, w.BaseModel.Client, &basemodel.TokenStore{Token: ""})
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
		if !w.filtering && (msg.String() != "q" || msg.String() != "ctrl+c") {
			w.companies, cmd = w.companies.Update(msg)
		}
		if w.filtering {
			switch msg.String() {
			case "esc":
				w.filtering = false
			case "enter":
				w.filtering = false
			}

			return w, nil
		} else {
			switch msg.String() {
			case "/":
				w.filtering = true
			case "s", "S":
				return w, func() tea.Msg {
					return messages.PageSwitchMsg{
						Page: messages.SearchPageNumber,
					}
				}
			case "enter":
				item := w.companies.SelectedItem()
				i := item.(companyItem)
				return w, func() tea.Msg {
					return messages.PageSwitchMsg{
						Page:    messages.CompanyPageNumber,
						Company: &i.company,
					}
				}

			case "r", "R":
				if len(w.companies.Items()) == 0 {
					return w, nil
				}

				item := w.companies.SelectedItem()
				i := item.(companyItem)
				err := w.removeCompanyFromWatchlist(i.company)
				if err != nil {
					return w, func() tea.Msg {
						return messages.PageSwitchMsg{
							Page: messages.ErrorPageNumber,
							Err:  err,
						}
					}
				}

				w.companies.RemoveItem(w.companies.Cursor())
				if w.companies.Cursor() == len(w.companies.Items()) {
					w.companies.CursorUp()
				}

			case "d", "D":
				if len(w.companies.Items()) == 0 {
					return w, nil
				}

				err := w.removeAllCompaniesFromWatchlist()
				if err != nil {
					return w, func() tea.Msg {
						return messages.PageSwitchMsg{
							Page: messages.ErrorPageNumber,
							Err:  err,
						}
					}
				}

				w.companies.SetItems([]list.Item{})

			case "p", "P":
				return w, func() tea.Msg {
					return messages.SmartPageSwitchMsg{
						Page: messages.ProfilePageNumber,
					}
				}

			case "b", "B":
				return w, func() tea.Msg {
					return messages.SmartPageSwitchMsg{
						Page: messages.BankRelationshipPageNumber,
					}
				}

			case "q", "ctrl+c":
				return w, func() tea.Msg {
					return messages.QuitMsg{}
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
		Foreground(cyan).
		Bold(true).
		Padding(0, 2).
		MarginBottom(1).
		Align(lipgloss.Center)

	header := "\n" + headerStyle.Render(w.titleBar) + "\n\n"

	if !w.loaded {
		return lipgloss.Place(w.BaseModel.Width, w.BaseModel.Height, lipgloss.Center, lipgloss.Center, w.spinner.View())
	}

	// Right panel (Top movers)
	colSymbol := lipgloss.NewStyle().Width(10)
	colPrice := lipgloss.NewStyle().Width(10)
	colChange := lipgloss.NewStyle().Width(10)
	colPercent := lipgloss.NewStyle().Width(10)

	title := lipgloss.NewStyle().
		Width(40).
		Align(lipgloss.Center).
		Render("Top Market Movers")

	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		colSymbol.Render("Ticker"),
		colPrice.Render("Price"),
		colChange.Render("Change"),
		colPercent.Render("% Change"),
	)

	rows := []string{
		title,
		"",
		headerRow,
	}

	for i, m := range w.movers {
		color := green
		if i >= 5 {
			color = red
		}

		price := fmt.Sprintf("%.2f", m.Price)
		change := fmt.Sprintf("%.2f", m.Change)
		percent := fmt.Sprintf("%.2f", m.PercentChange)

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			colSymbol.Render(m.Symbol),
			colPrice.Render(
				lipgloss.NewStyle().Foreground(color).Render(price),
			),
			colChange.Render(
				lipgloss.NewStyle().Foreground(color).Render(change),
			),
			colPercent.Render(
				lipgloss.NewStyle().Foreground(color).Render(percent),
			),
		)

		rows = append(rows, row)
	}

	right := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Padding(1, 2).
		Render(strings.Join(rows, "\n"))

	content := ""
	if !w.emptyWatchlist {
		content = lipgloss.JoinHorizontal(lipgloss.Top,
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

func (w WatchlistPage) removeCompanyFromWatchlist(company messages.CompanyInfo) error {
	_, err := requests.MakeRequest(http.MethodDelete, requests.BaseURL+"/watchlist/"+company.Symbol, nil, &http.Client{}, w.BaseModel.TokenStore)
	if err != nil {
		return err
	}

	return nil
}

func (w WatchlistPage) removeAllCompaniesFromWatchlist() error {
	_, err := requests.MakeRequest(http.MethodDelete, requests.BaseURL+"/watchlist", nil, &http.Client{}, w.BaseModel.TokenStore)
	if err != nil {
		return err
	}

	return nil
}

func (w *WatchlistPage) Reload() {
	w.companies.SetItems([]list.Item{})
	w.Reloaded = true
}
