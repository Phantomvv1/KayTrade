package searchpage

import (
	"encoding/json"
	"net/http"
	"strings"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SearchPage struct {
	BaseModel   basemodel.BaseModel
	searchField textinput.Model
	suggestions []Asset
	cursor      int
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

func NewSearchPage(client *http.Client) SearchPage {
	search := textinput.New()
	search.Placeholder = "Searching by symbol of the company"
	search.Width = 50
	search.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	search.Focus()

	return SearchPage{
		BaseModel:   basemodel.BaseModel{Client: client},
		searchField: search,
		name:        false,
	}
}

func (s SearchPage) Init() tea.Cmd {
	return textinput.Blink
}

func (s SearchPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg {
				return messages.PageSwitchWithoutInit{
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
			if s.cursor < len(s.suggestions)-1 {
				s.cursor++
			}
		case "ctrl+k", "up":
			if s.cursor > 0 {
				s.cursor--
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
					Page: messages.CompanyPageNumber,
					Comp: company,
				}
			}
		default:
			info, err := s.SendSearchRequest()
			if err != nil {
				return s, func() tea.Msg {
					return messages.PageSwitchMsg{
						Page: messages.ErrorPageNumber,
						Err:  err,
					}
				}
			}

			s.suggestions = info
		}
	}

	return s, nil
}

func (s SearchPage) View() string {
	return ""
}

func (s SearchPage) SendSearchRequest() ([]Asset, error) {
	arr := strings.Split(s.searchField.Placeholder, " ")
	body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/search?"+arr[2], nil, http.DefaultClient, s.BaseModel.Token)
	if err != nil {
		return nil, err
	}

	var response map[string][]Asset
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response["information"], nil
}

func (s SearchPage) GetCompanyInfo() (*CompanyInfo, error) {
	body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/company-information/"+s.suggestions[s.cursor].Symbol, nil, http.DefaultClient, s.BaseModel.Token)
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
	s.suggestions = nil
	s.cursor = 0
	s.name = false
}
