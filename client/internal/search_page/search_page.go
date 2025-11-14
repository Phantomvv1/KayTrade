package searchpage

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SearchPage struct {
	BaseModel   basemodel.BaseModel
	searchField textinput.Model
	suggestions list.Model
	name        bool
}

type Asset struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
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

type asset struct {
	asset Asset
}

func (c asset) Title() string       { return c.asset.Symbol }
func (c asset) Description() string { return c.asset.Name }
func (c asset) FilterValue() string { return c.asset.Symbol }

type itemMsg struct {
	items []list.Item
}

func NewSearchPage(client *http.Client) SearchPage {
	search := textinput.New()
	search.Placeholder = "Searching by symbol of the company"
	search.Width = 50
	search.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	search.Focus()

	delegate := list.NewDefaultDelegate()
	delegate.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("ctrl+k", "up"), key.WithHelp("ctrl+k/↑", "up")),
			key.NewBinding(key.WithKeys("ctrl+j", "down"), key.WithHelp("ctrl+j/↓", "down")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		}
	}

	delegate.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{}
	}

	sugg := list.New([]list.Item{}, delegate, 40, 30)
	sugg.Title = "Search results"
	sugg.KeyMap = list.KeyMap{}

	return SearchPage{
		BaseModel:   basemodel.BaseModel{Client: client},
		searchField: search,
		name:        false,
		suggestions: sugg,
	}
}

func (s SearchPage) Init() tea.Cmd {
	return textinput.Blink
}

func (s SearchPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.PageSwitchMsg:
		return s, func() tea.Msg {
			return msg
		}
	case itemMsg:
		s.suggestions.SetItems(nil)
		s.suggestions.SetItems(msg.items)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg {
				return messages.PageSwitchWithoutInitMsg{
					Page: messages.WatchlistPageNumber,
				}
			}
		case "tab":
			if !s.name {
				s.name = !s.name
				s.searchField.Placeholder = "Searching by name of the company"
			} else {
				s.name = !s.name
				s.searchField.Placeholder = "Searching by symbol of the company"
			}
		case "ctrl+j", "down":
			if s.suggestions.Cursor() < len(s.suggestions.Items())-1 {
				s.suggestions.Select(s.suggestions.Cursor() + 1)
			}
		case "ctrl+k", "up":
			if s.suggestions.Cursor() > 0 {
				s.suggestions.Select(s.suggestions.Cursor() - 1)
			}
		case "enter":
			company, err := s.GetCompanyInfo()
			if err != nil {
				return s, func() tea.Msg {
					return messages.PageSwitchMsg{
						Page: messages.ErrorPageNumber,
						Err:  err,
					}
				}
			}

			return s, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page:    messages.CompanyPageNumber,
					Company: company,
				}
			}
		default:
			if !s.name {
				if msg.String() != "backspace" {
					msg = tea.KeyMsg{Runes: []rune(strings.ToUpper(msg.String()))}
				}
			}

			old := s.searchField.Value()
			newField, cmd := s.searchField.Update(msg)
			s.searchField = newField

			if newField.Value() != old && newField.Value() != "" {
				return s, tea.Batch(s.SearchCmd(), cmd)
			}

			return s, cmd
		}
	}

	return s, nil
}

func (s SearchPage) View() string {
	cyan := lipgloss.Color("#00FFFF")
	purple := lipgloss.Color("#A020F0")

	title := lipgloss.NewStyle().
		Foreground(cyan).
		Bold(true).
		Padding(0, 2).
		MarginBottom(1).
		Align(lipgloss.Center).
		Render("🔎 Search")

	// Search bar block
	searchBox := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Width(s.BaseModel.Width / 2).
		Align(lipgloss.Left).
		Render(s.searchField.View())

	// Suggestions block
	suggestionsView := s.suggestions.View()
	suggestionsBox := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purple).
		Width(s.BaseModel.Width / 2).
		Height(s.BaseModel.Height - 10). // space for title + search
		Render(suggestionsView)

	// Put search bar above the suggestions:
	combined := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		searchBox,
		suggestionsBox,
	)

	// Center horizontally, 3 lines of padding at top
	return lipgloss.Place(
		s.BaseModel.Width,
		s.BaseModel.Height,
		lipgloss.Center,
		lipgloss.Top,
		"\n\n\n"+combined,
	)
}

func (s SearchPage) SendSearchRequest() ([]Asset, error) {
	arr := strings.Split(s.searchField.Placeholder, " ")
	value := strings.ReplaceAll(s.searchField.Value(), " ", "+")
	body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/search?"+arr[2]+"="+value, nil, http.DefaultClient, s.BaseModel.Token)
	if err != nil {
		return nil, err
	}

	var response map[string][]Asset
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	log.Println(response)
	return response["result"], nil
}

func (s SearchPage) GetCompanyInfo() (*CompanyInfo, error) {
	item := s.suggestions.Items()[s.suggestions.Cursor()].(asset)
	body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/company-information/"+item.asset.Symbol, nil, http.DefaultClient, s.BaseModel.Token)
	if err != nil {
		return nil, err
	}

	res := CompanyInfo{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *SearchPage) Reload() {
	s.searchField.SetValue("")
	s.suggestions.SetItems([]list.Item{})
	s.name = false
}

func (s SearchPage) SearchCmd() tea.Cmd {
	info, err := s.SendSearchRequest()
	if err != nil {
		return func() tea.Msg {
			return messages.PageSwitchMsg{
				Page: messages.ErrorPageNumber,
				Err:  err,
			}
		}
	}

	var res []list.Item
	for _, item := range info {
		res = append(res, asset{asset: item})
	}

	return func() tea.Msg {
		return itemMsg{
			items: res,
		}
	}
}
