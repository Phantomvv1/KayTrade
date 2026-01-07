package bankrelationshippage

import (
	"encoding/json"
	"net/http"
	"sync"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

type BanekRelationship struct {
	BaseModel         basemodel.BaseModel
	bankRelationships list.Model
	loaded            bool
	emptyBankRels     bool
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

func (b bankRelationshipItem) Title() string {
	if b.bankRelationship != nil {
		return b.bankRelationship.AccountNumber + " - " + "bank"
	} else if b.achRelationship != nil {
		return b.bankRelationship.AccountNumber + " - " + "ach"
	} else {
		return "Error no account number found"
	}
}
func (b bankRelationshipItem) Description() string {
	if b.bankRelationship != nil {
		return b.bankRelationship.BankCodeType + ", " + b.bankRelationship.Status
	} else if b.achRelationship != nil {
		return b.achRelationship.BankAccountType + ", " + b.achRelationship.Status
	} else {
		return "Error no details found"
	}
}
func (b bankRelationshipItem) FilterValue() string {
	if b.bankRelationship != nil {
		return b.bankRelationship.AccountNumber
	} else if b.achRelationship != nil {
		return b.bankRelationship.AccountNumber
	} else {
		return ""
	}
}

func NewBankRelationshipPage(client *http.Client, tokenStore *basemodel.TokenStore) BanekRelationship {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)

	return BanekRelationship{
		BaseModel:         basemodel.BaseModel{Client: client, TokenStore: tokenStore},
		bankRelationships: l,
		loaded:            false,
		emptyBankRels:     false,
	}
}

func (b BanekRelationship) Init() tea.Cmd {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var bankRelationships []bankRelationship
	var achRelationships []achRelationship
	var err1, err2 error

	go func() {
		defer wg.Done()
		body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"/funding/alpaca", nil, b.BaseModel.Client, b.BaseModel.TokenStore)
		if err != nil {
			err1 = err
			return
		}

		err1 = json.Unmarshal(body, &bankRelationships)
	}()

	go func() {
		defer wg.Done()
		body, err := requests.MakeRequest(http.MethodGet, requests.BaseURL+"funding/ach", nil, b.BaseModel.Client, b.BaseModel.TokenStore)
		if err != nil {
			err2 = err
			return
		}
		err2 = json.Unmarshal(body, &achRelationships)
	}()

	wg.Wait()

	if err1 != nil || err2 != nil {
		if err1 != nil {
			return func() tea.Msg {
				return initResult{err: err1}
			}
		} else {
			return func() tea.Msg {
				return initResult{err: err2}
			}
		}
	}

	return func() tea.Msg {
		return initResult{bankRelationships: bankRelationships, achRelationships: achRelationships, err: nil}
	}
}

func (b BanekRelationship) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
	case initResult:
		b.loaded = true
		for i, bankRel := range msg.bankRelationships {
			b.bankRelationships.InsertItem(i, bankRelationshipItem{bankRelationship: &bankRel})
		}

		for i, bankRel := range msg.achRelationships {
			b.bankRelationships.InsertItem(i, bankRelationshipItem{achRelationship: &bankRel})
		}

		b.bankRelationships.SetSize(b.BaseModel.Width/2-4, b.BaseModel.Height-8)

		if len(b.bankRelationships.Items()) == 0 {
			b.emptyBankRels = true
		}
	}

	return b, nil
}

func (b BanekRelationship) View() string {
	return ""
}
