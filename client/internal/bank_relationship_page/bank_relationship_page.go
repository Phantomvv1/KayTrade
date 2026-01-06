package bankrelationshippage

import (
	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	normalBankType = "bank"
	bankTypeAch    = "ach"
	inputWidth     = 27
)

type bankInputs struct {
	name          textinput.Model
	bankCode      textinput.Model
	bankCodeType  []string
	accountNumber textinput.Model
	country       textinput.Model // reqiured if BCT = "BIC"
	stateProvince textinput.Model // reqiured if BCT = "BIC"
	city          textinput.Model // reqiured if BCT = "BIC"
	streetAddress textinput.Model // reqiured if BCT = "BIC"
}

type achInputs struct {
	accountOwnerName  textinput.Model
	bankAccountType   []string
	bankAccountNumber textinput.Model
	bankRoutingNumber textinput.Model
	nickname          textinput.Model // not reqiured
}

type BankRelationship struct {
	BaseModel  basemodel.BaseModel
	bankType   string
	bankInputs bankInputs
	achInputs  achInputs
}

func NewBankInputs() bankInputs {
	name := textinput.New()
	name.Placeholder = "Name"
	name.Width = inputWidth
	name.CharLimit = 50

	bankCode := textinput.New()
	bankCode.Placeholder = "Bank code"
	bankCode.Width = inputWidth
	bankCode.CharLimit = 20

	accountNumber := textinput.New()
	accountNumber.Placeholder = "Account number"
	accountNumber.Width = inputWidth
	accountNumber.CharLimit = 60

	country := textinput.New()
	country.Placeholder = "Country"
	country.Width = inputWidth
	country.CharLimit = 10

	stateProvince := textinput.New()
	stateProvince.Placeholder = "State/Province"
	stateProvince.Width = inputWidth
	stateProvince.CharLimit = 30

	city := textinput.New()
	city.Placeholder = "City"
	city.Width = inputWidth
	city.CharLimit = 20

	streetAddress := textinput.New()
	streetAddress.Placeholder = "Street address"
	streetAddress.Width = inputWidth
	streetAddress.CharLimit = 15

	return bankInputs{
		name:          name,
		bankCode:      bankCode,
		bankCodeType:  []string{"ABA", "BIC"}, // ABA - domestic, BIC - international
		accountNumber: accountNumber,
		country:       country,
		stateProvince: stateProvince,
		city:          city,
		streetAddress: streetAddress,
	}
}

func NewAchInputs() achInputs {
	accountOwnerName := textinput.New()
	accountOwnerName.Placeholder = "Account owner name"
	accountOwnerName.Width = inputWidth
	accountOwnerName.CharLimit = 50

	bankAccountNumber := textinput.New()
	bankAccountNumber.Placeholder = "Bank account number"
	bankAccountNumber.Width = inputWidth
	bankAccountNumber.CharLimit = 60

	bankRoutingNumber := textinput.New()
	bankRoutingNumber.Placeholder = "Bank routing number"
	bankRoutingNumber.Width = inputWidth
	bankRoutingNumber.CharLimit = 10

	nickname := textinput.New()
	nickname.Placeholder = "Nickname"
	nickname.Width = inputWidth
	nickname.CharLimit = 30

	return achInputs{
		accountOwnerName:  accountOwnerName,
		bankAccountType:   []string{"CHECKING", "SAVINGS"},
		bankAccountNumber: bankAccountNumber,
		bankRoutingNumber: bankRoutingNumber,
		nickname:          nickname,
	}
}

func (b BankRelationship) Init() tea.Cmd {
	return nil
}
func (b BankRelationship) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return b, nil
}

func (b BankRelationship) View() string {
	return ""
}
