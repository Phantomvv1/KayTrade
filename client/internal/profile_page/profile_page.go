package profilepage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TradingDetails struct {
	AccountNumber       string `json:"account_number"`
	AccruedFees         string `json:"accrued_fees"`
	BuyingPower         string `json:"buying_power"`
	Cash                string `json:"cash"`
	CashTransferable    string `json:"cash_transferable"`
	CashWithdrawable    string `json:"cash_withdrawable"`
	Currency            string `json:"currency"`
	Equity              string `json:"equity"`
	IntradayAdjustments string `json:"intraday_adjustments"`
	InitialMargin       string `json:"initial_margin"`
	Status              string `json:"status"`
}

type Contact struct {
	City          string   `json:"city"`
	Country       string   `json:"country"`
	EmailAddress  string   `json:"email_address"`
	PhoneNumber   string   `json:"phone_number"`
	PostalCode    string   `json:"postal_code"`
	State         string   `json:"state"`
	StreetAddress []string `json:"street_address"`
	Unit          string   `json:"unit"`
}

type Identity struct {
	CountryOfTaxResidence string   `json:"country_of_tax_residence"`
	DateOfBirth           string   `json:"date_of_birth"`
	FamilyName            string   `json:"family_name"`
	FundingSource         []string `json:"funding_source"`
	GivenName             string   `json:"given_name"`
	PartyType             string   `json:"party_type"`
	TaxIdType             string   `json:"tax_id_type"`
}

type TrustedContact struct {
	EmailAddress string `json:"email_address"`
	FamilyName   string `json:"family_name"`
	GivenName    string `json:"given_name"`
}

type AlpacaAccount struct {
	AccountType    string         `json:"account_type"`
	Contact        Contact        `json:"contact"`
	CreatedAt      string         `json:"created_at"`
	CryptoStatus   string         `json:"crypto_status"`
	Currency       string         `json:"currency"`
	EnabledAssets  []string       `json:"enabled_assets"`
	Identity       Identity       `json:"identity"`
	LastEquity     string         `json:"last_equity"`
	Status         string         `json:"status"`
	TradingType    string         `json:"trading_type"`
	TrustedContact TrustedContact `json:"trusted_contact"`
}

type ProfilePage struct {
	BaseModel      basemodel.BaseModel
	tradingDetails TradingDetails
	alpacaAccount  AlpacaAccount
	loading        bool
	Reloaded       bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1)

	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#BB88FF")).
				Bold(true).
				Underline(true).
				MarginTop(1).
				MarginBottom(1).
				Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB88FF")).
			Width(25)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true)

	statusActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(1, 0)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BB88FF")).
			Padding(1, 2).
			MarginBottom(1)
)

type profileDataMsg struct {
	tradingDetails TradingDetails
	alpacaAccount  AlpacaAccount
	err            error
}

func NewProfilePage(client *http.Client) ProfilePage {
	return ProfilePage{
		BaseModel: basemodel.BaseModel{Client: client},
		loading:   true,
	}
}

func (p ProfilePage) Init() tea.Cmd {
	return p.fetchProfileData
}

func (p ProfilePage) fetchProfileData() tea.Msg {
	var tradingDetails TradingDetails
	var alpacaAccount AlpacaAccount

	var err1, err2 error
	go func() {
		body, err := requests.MakeRequest(
			http.MethodGet,
			requests.BaseURL+"/users/trading-details",
			nil,
			p.BaseModel.Client,
			p.BaseModel.Token,
		)
		if err != nil {
			err1 = err
			return
		}

		if err := json.Unmarshal(body, &tradingDetails); err != nil {
			err1 = fmt.Errorf("failed to parse trading details: %v", err)
			return
		}
	}()

	go func() {
		body, err := requests.MakeRequest(
			http.MethodGet,
			requests.BaseURL+"/users/alpaca?status=all",
			nil,
			p.BaseModel.Client,
			p.BaseModel.Token,
		)
		if err != nil {
			err2 = err
			return
		}

		if err := json.Unmarshal(body, &alpacaAccount); err != nil {
			err2 = fmt.Errorf("failed to parse account details: %v", err)
			return
		}
	}()

	go func() {
		body, err := requests.MakeRequest(
			http.MethodGet,
			requests.BaseURL+"/trading/alpaca?status=all",
			nil,
			p.BaseModel.Client,
			p.BaseModel.Token,
		)
		if err != nil {
			err2 = err
			return
		}

		if err := json.Unmarshal(body, &alpacaAccount); err != nil {
			err2 = fmt.Errorf("failed to parse account details: %v", err)
			return
		}
	}()

	if err1 != nil {
		return profileDataMsg{err: err1}
	}

	if err2 != nil {
		return profileDataMsg{err: err2}
	}

	return profileDataMsg{
		tradingDetails: tradingDetails,
		alpacaAccount:  alpacaAccount,
	}
}

func (p ProfilePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return p, tea.Quit
		case "esc":
			return p, func() tea.Msg {
				return messages.PageSwitchMsg{Page: messages.WatchlistPageNumber}
			}
		}

	case profileDataMsg:
		p.loading = false
		if msg.err != nil {
			return p, func() tea.Msg {
				return messages.PageSwitchMsg{
					Page: messages.ErrorPageNumber,
					Err:  msg.err,
				}
			}
		} else {
			p.tradingDetails = msg.tradingDetails
			p.alpacaAccount = msg.alpacaAccount
		}
		return p, nil

	}

	return p, nil
}

func (p ProfilePage) View() string {
	if p.loading {
		return lipgloss.Place(
			p.BaseModel.Width,
			p.BaseModel.Height,
			lipgloss.Center,
			lipgloss.Center,
			"Loading profile data...",
		)
	}

	title := titleStyle.Render("👤 Profile")

	personalInfo := p.renderPersonalInfo()

	tradingAccount := p.renderTradingAccount()

	contactInfo := p.renderContactInfo()

	accountSettings := p.renderAccountSettings()

	leftColumn := lipgloss.JoinVertical(lipgloss.Left, personalInfo, contactInfo)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, tradingAccount, accountSettings)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, "  ", rightColumn)

	help := helpStyle.Render("r: refresh • esc: back • q: quit")

	finalView := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		content,
		"",
		help,
	)

	return lipgloss.Place(
		p.BaseModel.Width,
		p.BaseModel.Height,
		lipgloss.Center,
		lipgloss.Center,
		finalView,
	)
}

func (p ProfilePage) renderPersonalInfo() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("📋 Personal Information"))

	fullName := fmt.Sprintf("%s %s", p.alpacaAccount.Identity.GivenName, p.alpacaAccount.Identity.FamilyName)
	rows = append(rows, p.renderField("Name", fullName))
	rows = append(rows, p.renderField("Date of Birth", p.alpacaAccount.Identity.DateOfBirth))
	rows = append(rows, p.renderField("Email", p.alpacaAccount.Contact.EmailAddress))
	rows = append(rows, p.renderField("Phone", p.alpacaAccount.Contact.PhoneNumber))
	rows = append(rows, p.renderField("Tax Country", p.alpacaAccount.Identity.CountryOfTaxResidence))

	if len(p.alpacaAccount.Identity.FundingSource) > 0 {
		rows = append(rows, p.renderField("Funding Source", strings.Join(p.alpacaAccount.Identity.FundingSource, ", ")))
	}

	return boxStyle.Width(50).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (p ProfilePage) renderTradingAccount() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("💰 Trading Account"))

	statusValue := statusActiveStyle.Render(p.tradingDetails.Status)
	rows = append(rows, labelStyle.Render("Status:")+"  "+statusValue)

	rows = append(rows, p.renderField("Equity", "$"+p.tradingDetails.Equity))
	rows = append(rows, p.renderField("Cash", "$"+p.tradingDetails.Cash))
	rows = append(rows, p.renderField("Buying Power", "$"+p.tradingDetails.BuyingPower))
	rows = append(rows, p.renderField("Cash Withdrawable", "$"+p.tradingDetails.CashWithdrawable))
	rows = append(rows, p.renderField("Cash Transferable", "$"+p.tradingDetails.CashTransferable))
	rows = append(rows, p.renderField("Initial Margin", "$"+p.tradingDetails.InitialMargin))
	rows = append(rows, p.renderField("Accrued Fees", "$"+p.tradingDetails.AccruedFees))
	rows = append(rows, p.renderField("Currency", p.tradingDetails.Currency))

	return boxStyle.Width(50).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (p ProfilePage) renderContactInfo() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("📍 Address"))

	if len(p.alpacaAccount.Contact.StreetAddress) > 0 {
		rows = append(rows, p.renderField("Street", p.alpacaAccount.Contact.StreetAddress[0]))
	}
	if p.alpacaAccount.Contact.Unit != "" {
		rows = append(rows, p.renderField("Unit", p.alpacaAccount.Contact.Unit))
	}
	cityStateZip := fmt.Sprintf("%s, %s %s",
		p.alpacaAccount.Contact.City,
		p.alpacaAccount.Contact.State,
		p.alpacaAccount.Contact.PostalCode,
	)
	rows = append(rows, p.renderField("City", cityStateZip))
	rows = append(rows, p.renderField("Country", p.alpacaAccount.Contact.Country))

	// Trusted Contact
	if p.alpacaAccount.TrustedContact.GivenName != "" {
		rows = append(rows, "")
		rows = append(rows, sectionTitleStyle.Render("🤝 Trusted Contact"))
		trustedName := fmt.Sprintf("%s %s",
			p.alpacaAccount.TrustedContact.GivenName,
			p.alpacaAccount.TrustedContact.FamilyName,
		)
		rows = append(rows, p.renderField("Name", trustedName))
		rows = append(rows, p.renderField("Email", p.alpacaAccount.TrustedContact.EmailAddress))
	}

	return boxStyle.Width(50).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (p ProfilePage) renderAccountSettings() string {
	var rows []string

	rows = append(rows, sectionTitleStyle.Render("⚙️  Account Settings"))

	rows = append(rows, p.renderField("Account Type", strings.ToUpper(p.alpacaAccount.AccountType)))
	rows = append(rows, p.renderField("Trading Type", strings.ToUpper(p.alpacaAccount.TradingType)))

	cryptoStatus := p.alpacaAccount.CryptoStatus
	if cryptoStatus == "ACTIVE" {
		cryptoStatus = statusActiveStyle.Render(cryptoStatus)
		rows = append(rows, labelStyle.Render("Crypto Status:")+"  "+cryptoStatus)
	} else {
		rows = append(rows, p.renderField("Crypto Status", cryptoStatus))
	}

	if len(p.alpacaAccount.EnabledAssets) > 0 {
		assets := strings.Join(p.alpacaAccount.EnabledAssets, ", ")
		rows = append(rows, p.renderField("Enabled Assets", strings.ToUpper(assets)))
	}

	// Parse and format created date
	if createdAt, err := time.Parse(time.RFC3339, p.alpacaAccount.CreatedAt); err == nil {
		formatted := createdAt.Format("Jan 02, 2006")
		rows = append(rows, p.renderField("Member Since", formatted))
	}

	return boxStyle.Width(50).Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (p ProfilePage) renderField(label, value string) string {
	return labelStyle.Render(label+":") + "  " + valueStyle.Render(value)
}

func (p *ProfilePage) Reload() {
	p.alpacaAccount = AlpacaAccount{}
	p.tradingDetails = TradingDetails{}
	p.Reloaded = true
}
