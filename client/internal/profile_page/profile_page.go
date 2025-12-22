package profilepage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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

type Order struct {
	AssetClass     string  `json:"asset_class"`
	AssetID        string  `json:"asset_id"`
	CanceledAt     string  `json:"canceled_at"`
	CreatedAt      string  `json:"created_at"`
	ExpiredAt      string  `json:"expired_at"`
	ExpiresAt      string  `json:"expires_at"`
	FailedAt       string  `json:"failed_at"`
	FilledAt       string  `json:"filled_at"`
	FilledAvgPrice string  `json:"filled_avg_price"`
	FilledQty      string  `json:"filled_qty"`
	ID             string  `json:"id"`
	LimitPrice     *string `json:"limit_price"`
	Notional       *string `json:"notional"`
	OrderType      string  `json:"order_type"`
	PositionIntent string  `json:"position_intent"`
	Quantity       string  `json:"qty"`
	Side           string  `json:"side"`
	Status         string  `json:"status"`
	StopPrice      *string `json:"stop_price"`
	SubmittedAt    string  `json:"submitted_at"`
	Symbol         string  `json:"symbol"`
	TimeInForce    string  `json:"time_in_force"`
	TrailPercent   *string `json:"trail_percent"`
	TrailPrice     *string `json:"trail_price"`
	Type           string  `json:"type"`
	UpdatedAt      string  `json:"updated_at"`
}

type Position struct {
	AssetClass             string `json:"asset_class"`
	AssetID                string `json:"asset_id"`
	AssetMarginable        bool   `json:"asset_marginable"`
	AvgEntryPrice          string `json:"avg_entry_price"`
	ChangeToday            string `json:"change_today"`
	CostBasis              string `json:"cost_basis"`
	CurrentPrice           string `json:"current_price"`
	Exchange               string `json:"exchange"`
	LastdayPrice           string `json:"lastday_price"`
	MarketValue            string `json:"market_value"`
	Qty                    string `json:"qty"`
	QtyAvailable           string `json:"qty_available"`
	Side                   string `json:"side"`
	Symbol                 string `json:"symbol"`
	UnrealizedIntradayPL   string `json:"unrealized_intraday_pl"`
	UnrealizedIntradayPLPC string `json:"unrealized_intraday_plpc"`
	UnrealizedPL           string `json:"unrealized_pl"`
	UnrealizedPLPC         string `json:"unrealized_plpc"`
}

type ProfilePage struct {
	BaseModel      basemodel.BaseModel
	tradingDetails TradingDetails
	alpacaAccount  AlpacaAccount
	orders         list.Model
	positions      list.Model
	filtering      bool
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

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BB88FF")).
			Padding(1, 2).
			MarginBottom(1)

	activeListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FFFF")).
			Padding(0, 1)

	inactiveListStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#666666")).
				Padding(0, 1)
)

type profileDataMsg struct {
	tradingDetails TradingDetails
	alpacaAccount  AlpacaAccount
	orders         []Order
	positions      []Position
	err            error
}

type orderItem struct {
	order Order
}

func (o orderItem) Title() string {
	return o.order.Quantity + "x " + o.order.Symbol + " - " + o.order.CreatedAt
}

func (o orderItem) Description() string {
	if o.order.CanceledAt != "" {
		return o.order.Side + ", Canceled at: " + o.order.CanceledAt
	}

	if o.order.FilledAt == "" {
		return o.order.Side + ", Expires at: " + o.order.ExpiresAt
	} else {
		return o.order.Side + ", Filled at:" + o.order.FilledAt
	}
}

func (o orderItem) FilterValue() string { return o.order.Symbol }

type positionItem struct {
	position Position
}

func (p positionItem) Title() string {
	return p.position.Qty + "x " + p.position.Symbol
}

func (p positionItem) Description() string {
	return "Bought for: " + p.position.CostBasis + ", Price: " + p.position.CurrentPrice
}

func (p positionItem) FilterValue() string { return p.position.Symbol }

func NewProfilePage(client *http.Client) ProfilePage {
	ordersList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	ordersList.FilterInput.Focus()
	ordersList.Title = "Orders"
	ordersList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("ctrl+h", "ctrl+left"), key.WithHelp("ctrl+h/←", "switch list")),
			key.NewBinding(key.WithKeys("ctrl+l", "ctrl+right"), key.WithHelp("ctrl+l/→", "switch list")),
			key.NewBinding(key.WithKeys("s", "S"), key.WithHelp("s (sell)", "position")),
		}
	}

	positionsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	positionsList.FilterInput.Blur()
	positionsList.Title = "Positions"
	positionsList.SetShowHelp(false)

	return ProfilePage{
		BaseModel: basemodel.BaseModel{Client: client},
		orders:    ordersList,
		positions: positionsList,
		filtering: false,
		loading:   true,
		Reloaded:  true,
	}
}

func (p ProfilePage) Init() tea.Cmd {
	return p.fetchProfileData
}

func (p ProfilePage) fetchProfileData() tea.Msg {
	tradingDetails := TradingDetails{}
	alpacaAccount := AlpacaAccount{}
	orders := []Order{}
	positions := []Position{}

	wg := sync.WaitGroup{}
	wg.Add(4)
	var err1, err2, err3, err4 error
	go func() {
		defer wg.Done()
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
		defer wg.Done()

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
		defer wg.Done()

		body, err := requests.MakeRequest(
			http.MethodGet,
			requests.BaseURL+"/trading/alpaca?status=all",
			nil,
			p.BaseModel.Client,
			p.BaseModel.Token,
		)
		if err != nil {
			err3 = err
			return
		}

		if err := json.Unmarshal(body, &orders); err != nil {
			err3 = fmt.Errorf("failed to parse orders: %v", err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		body, err := requests.MakeRequest(
			http.MethodGet,
			requests.BaseURL+"/trading/positions",
			nil,
			p.BaseModel.Client,
			p.BaseModel.Token,
		)
		if err != nil {
			err4 = err
			return
		}

		if err := json.Unmarshal(body, &positions); err != nil {
			err4 = fmt.Errorf("failed to parse positions: %v", err)
			return
		}
	}()

	wg.Wait()

	if err1 != nil {
		return profileDataMsg{err: err1}
	}

	if err2 != nil {
		return profileDataMsg{err: err2}
	}

	if err3 != nil {
		return profileDataMsg{err: err3}
	}

	if err4 != nil {
		return profileDataMsg{err: err4}
	}

	return profileDataMsg{
		tradingDetails: tradingDetails,
		alpacaAccount:  alpacaAccount,
		orders:         orders,
		positions:      positions,
	}
}

func (p ProfilePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if p.filtering {
			switch msg.String() {
			case "enter", "esc":
				p.filtering = false
				var cmd tea.Cmd
				if p.orders.FilterInput.Focused() {
					p.orders, cmd = p.orders.Update(msg)
				} else {
					p.positions, cmd = p.positions.Update(msg)
				}
				return p, cmd

			default:
				var cmd tea.Cmd
				if p.orders.FilterInput.Focused() {
					p.orders, cmd = p.orders.Update(msg)
				} else {
					p.positions, cmd = p.positions.Update(msg)
				}
				return p, cmd
			}
		} else {
			switch msg.String() {
			case "q", "ctrl+c":
				return p, tea.Quit

			case "esc":
				activeFilterValue := ""
				if p.orders.FilterInput.Focused() {
					activeFilterValue = p.orders.FilterValue()
				} else {
					activeFilterValue = p.positions.FilterValue()
				}

				if activeFilterValue != "" {
					var cmd tea.Cmd
					if p.orders.FilterInput.Focused() {
						p.orders, cmd = p.orders.Update(msg)
					} else {
						p.positions, cmd = p.positions.Update(msg)
					}
					return p, cmd
				}

				return p, func() tea.Msg {
					return messages.SmartPageSwitchMsg{Page: messages.WatchlistPageNumber}
				}

			case "ctrl+h", "ctrl+left":
				p.orders.FilterInput.Focus()
				p.positions.FilterInput.Blur()
				return p, nil

			case "ctrl+l", "ctrl+right":
				p.positions.FilterInput.Focus()
				p.orders.FilterInput.Blur()
				return p, nil

			case "/":
				p.filtering = true
				var cmd tea.Cmd
				if p.orders.FilterInput.Focused() {
					p.orders, cmd = p.orders.Update(msg)
				} else {
					p.positions, cmd = p.positions.Update(msg)
				}
				return p, cmd

			case "s", "S":
				if p.positions.FilterInput.Focused() {
					position := p.positions.Items()[p.positions.Cursor()].(positionItem)
					maxQuantity, err := strconv.ParseFloat(position.position.Qty, 64)
					if err != nil {
						return p, func() tea.Msg {
							return messages.PageSwitchMsg{
								Page: messages.ErrorPageNumber,
								Err:  err,
							}
						}
					}

					return p, func() tea.Msg {
						return messages.PageSwitchMsg{
							Page:        messages.SellPageNumber,
							Symbol:      position.position.Symbol,
							MaxQuantity: maxQuantity,
						}
					}
				}

			default:
				var cmd tea.Cmd
				if p.orders.FilterInput.Focused() {
					p.orders, cmd = p.orders.Update(msg)
				} else {
					p.positions, cmd = p.positions.Update(msg)
				}
				return p, cmd
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
			for i, order := range msg.orders {
				p.orders.InsertItem(i, orderItem{order: order})
			}

			for i, position := range msg.positions {
				p.positions.InsertItem(i, positionItem{position: position})
			}
		}

		p.orders.SetSize(p.BaseModel.Width/2-10, p.BaseModel.Height-16)
		p.positions.SetSize(p.BaseModel.Width/3-30, p.BaseModel.Height-16)

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

	// FIX: Temporary fix
	title := titleStyle.Render("👤 Profile")
	// centeredTitle := lipgloss.Place(p.BaseModel.Width, lipgloss.Height(title), lipgloss.Center, lipgloss.Top, title)

	personalInfo := p.renderPersonalInfo()

	tradingAccount := p.renderTradingAccount()

	contactInfo := p.renderContactInfo()

	accountSettings := p.renderAccountSettings()

	leftInfoColumn := lipgloss.JoinVertical(lipgloss.Left, personalInfo, contactInfo)
	rightInfoColumn := lipgloss.JoinVertical(lipgloss.Left, tradingAccount, accountSettings)

	infoColumns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftInfoColumn,
		lipgloss.NewStyle().MarginLeft(1).Render(rightInfoColumn),
	)

	var ordersView, positionsView string
	if p.orders.FilterInput.Focused() {
		ordersView = activeListStyle.Render(p.orders.View())
		positionsView = inactiveListStyle.Render(p.positions.View())
	} else {
		ordersView = inactiveListStyle.Render(p.orders.View())
		positionsView = activeListStyle.Render(p.positions.View())
	}

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		ordersView,
		lipgloss.NewStyle().MarginLeft(1).MarginRight(1).Render(infoColumns),
		positionsView,
	)

	finalView := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		// centeredTitle,
		title,
		"",
		content,
	)

	return finalView
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
	p.loading = true
	p.Reloaded = true
}
