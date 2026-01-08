package bankrelationshippage

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type bankRelationship struct {
	AccountID     string `json:"account_id"`
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	BankCodeType  string `json:"bank_code_type"`
	City          string `json:"city"`
	Country       string `json:"country"`
	CreatedAt     string `json:"created_at"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	PostalCode    string `json:"postal_code"`
	StateProvince string `json:"state_province"`
	Status        string `json:"status"`
	StreetAddress string `json:"street_address"`
	UpdatedAt     string `json:"updated_at"`
}

type achRelationship struct {
	AccountID         string  `json:"account_id"`
	AccountOwnerName  string  `json:"account_owner_name"`
	BankAccountNumber string  `json:"bank_account_number"`
	BankAccountType   string  `json:"bank_account_type"`
	BankRoutingNumber string  `json:"bank_routing_number"`
	CreatedAt         string  `json:"created_at"`
	ID                string  `json:"id"`
	Nickname          *string `json:"nickname"`
	ProcessorToken    *string `json:"processor_token"`
	Status            string  `json:"status"`
	UpdatedAt         string  `json:"updated_at"`
}

type BankRelationshipPage struct {
	BaseModel         basemodel.BaseModel
	bankRelationships list.Model
	loaded            bool
	emptyBankRels     bool
	Reloaded          bool
}

type initResult struct {
	bankRelationships []bankRelationship
	achRelationships  []achRelationship
	err               error
}

type bankRelationshipItem struct {
	bankRelationship *bankRelationship
	achRelationship  *achRelationship
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true)

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB88FF")).
			Italic(true).
			Padding(2, 0)
)

func (b bankRelationshipItem) Title() string {
	if b.bankRelationship != nil {
		return b.bankRelationship.AccountNumber + " - Bank"
	} else if b.achRelationship != nil {
		return b.achRelationship.BankAccountNumber + " - ACH"
	} else {
		return "Error: No account number found"
	}
}

func (b bankRelationshipItem) Description() string {
	if b.bankRelationship != nil {
		return b.bankRelationship.BankCodeType + ", " + b.bankRelationship.Status
	} else if b.achRelationship != nil {
		return b.achRelationship.BankAccountType + ", " + b.achRelationship.Status
	} else {
		return "Error: No details found"
	}
}

func (b bankRelationshipItem) FilterValue() string {
	if b.bankRelationship != nil {
		return b.bankRelationship.AccountNumber
	} else if b.achRelationship != nil {
		return b.achRelationship.BankAccountNumber
	} else {
		return ""
	}
}

func NewBankRelationshipPage(client *http.Client, tokenStore *basemodel.TokenStore) BankRelationshipPage {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Bank Relationships"
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("c", "C"), key.WithHelp("c (create)", "relationship")),
		}
	}

	return BankRelationshipPage{
		BaseModel:         basemodel.BaseModel{Client: client, TokenStore: tokenStore},
		bankRelationships: l,
		loaded:            false,
		emptyBankRels:     false,
		Reloaded:          true,
	}
}

func (b BankRelationshipPage) Init() tea.Cmd {
	return b.fetchBankRelationships
}

func (b BankRelationshipPage) fetchBankRelationships() tea.Msg {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var bankRelationships []bankRelationship
	var achRelationships []achRelationship
	var err1, err2 error

	go func() {
		defer wg.Done()
		body, err := requests.MakeRequest(
			http.MethodGet,
			requests.BaseURL+"/funding/alpaca",
			nil,
			b.BaseModel.Client,
			b.BaseModel.TokenStore,
		)
		if err != nil {
			err1 = err
			return
		}

		err1 = json.Unmarshal(body, &bankRelationships)
	}()

	go func() {
		defer wg.Done()
		body, err := requests.MakeRequest(
			http.MethodGet,
			requests.BaseURL+"/funding/ach",
			nil,
			b.BaseModel.Client,
			b.BaseModel.TokenStore,
		)
		if err != nil {
			err2 = err
			return
		}
		err2 = json.Unmarshal(body, &achRelationships)
	}()

	wg.Wait()

	if err1 != nil {
		return initResult{err: err1}
	}

	if err2 != nil {
		return initResult{err: err2}
	}

	return initResult{
		bankRelationships: bankRelationships,
		achRelationships:  achRelationships,
		err:               nil,
	}
}

func (b BankRelationshipPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return b, func() tea.Msg {
				return messages.QuitMsg{}
			}

		case "esc":
			if b.bankRelationships.FilterValue() != "" {
				var cmd tea.Cmd
				b.bankRelationships, cmd = b.bankRelationships.Update(msg)
				return b, cmd
			}

			return b, func() tea.Msg {
				return messages.SmartPageSwitchMsg{
					Page: messages.WatchlistPageNumber,
				}
			}

		case "c", "C":
			return b, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.BankRelationshipCreationPageNumber,
				}
			}

		case "f", "F":
			bank := b.bankRelationships.SelectedItem().(bankRelationshipItem)
			relType := strings.Split(bank.Title(), " - ")[1]
			switch relType {
			case "ach":
				return b, func() tea.Msg {
					return messages.PageSwitchMsg{
						Page: messages.TransfersPageNumber,
						FundingInformation: &messages.FundingInformation{
							TransferType:   relType,
							RelationshipId: bank.achRelationship.ID,
						},
					}
				}

			case "bank":
				return b, func() tea.Msg {
					return messages.PageSwitchMsg{
						Page: messages.TransfersPageNumber,
						FundingInformation: &messages.FundingInformation{
							TransferType: relType,
							BankId:       bank.bankRelationship.ID,
						},
					}
				}
			}

		default:
			var cmd tea.Cmd
			b.bankRelationships, cmd = b.bankRelationships.Update(msg)
			return b, cmd
		}

	case initResult:
		if msg.err != nil {
			return b, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.ErrorPageNumber,
					Err:  msg.err,
				}
			}
		}

		b.loaded = true

		idx := 0
		for _, rel := range msg.bankRelationships {
			b.bankRelationships.InsertItem(idx, bankRelationshipItem{bankRelationship: &rel})
			idx++
		}

		for _, rel := range msg.achRelationships {
			b.bankRelationships.InsertItem(idx, bankRelationshipItem{achRelationship: &rel})
			idx++
		}

		b.bankRelationships.SetSize(b.BaseModel.Width/2-4, b.BaseModel.Height-8)

		if len(b.bankRelationships.Items()) == 0 {
			b.emptyBankRels = true
		}

		return b, nil
	}

	var cmd tea.Cmd
	b.bankRelationships, cmd = b.bankRelationships.Update(msg)
	return b, cmd
}

func (b BankRelationshipPage) View() string {
	if !b.loaded {
		return lipgloss.Place(
			b.BaseModel.Width,
			b.BaseModel.Height,
			lipgloss.Center,
			lipgloss.Center,
			"Loading bank relationships...",
		)
	}

	title := titleStyle.Render("🏦 Bank Relationships")

	var content string
	if b.emptyBankRels {
		emptyMessage := emptyStyle.Render("No bank relationships found.\nAdd a bank account to start trading.")
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			title,
			"",
			"",
			emptyMessage,
		)
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			title,
			"\n\n\n\n",
			b.bankRelationships.View(),
		)
	}

	finalView := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.Place(b.BaseModel.Width, lipgloss.Height(content), lipgloss.Center, lipgloss.Top, content),
		"",
	)

	return lipgloss.Place(
		b.BaseModel.Width,
		b.BaseModel.Height,
		lipgloss.Center,
		lipgloss.Top,
		finalView,
	)
}

func (b *BankRelationshipPage) Reload() {
	b.bankRelationships.SetItems([]list.Item{})
	b.loaded = false
	b.emptyBankRels = false
	b.Reloaded = true
}
