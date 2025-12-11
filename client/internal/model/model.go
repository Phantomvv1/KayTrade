package model

import (
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"

	buypage "github.com/Phantomvv1/KayTrade/internal/buy_page"
	companypage "github.com/Phantomvv1/KayTrade/internal/company_page"
	errorpage "github.com/Phantomvv1/KayTrade/internal/error_page"
	landingpage "github.com/Phantomvv1/KayTrade/internal/landing_page"
	loginpage "github.com/Phantomvv1/KayTrade/internal/login_page"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	searchpage "github.com/Phantomvv1/KayTrade/internal/search_page"
	tradinginfopage "github.com/Phantomvv1/KayTrade/internal/trading_info_page"
	watchlistpage "github.com/Phantomvv1/KayTrade/internal/watchlist_page"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	landingPage     landingpage.LandingPage
	errorPage       errorpage.ErrorPage
	watchlistPage   watchlistpage.WatchlistPage
	loginPage       loginpage.LoginPage
	searchPage      searchpage.SearchPage
	companyPage     companypage.CompanyPage
	buyPage         buypage.BuyPage
	tradingInfoPage tradinginfopage.TradingInfoPage
	currentPage     int
}

func NewModel() Model {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Println(err)
	}

	client := &http.Client{Jar: jar}

	return Model{
		landingPage:     landingpage.LandingPage{},
		errorPage:       errorpage.ErrorPage{},
		watchlistPage:   watchlistpage.NewWatchlistPage(client),
		loginPage:       loginpage.NewLoginPage(client),
		searchPage:      searchpage.NewSearchPage(client),
		companyPage:     companypage.NewCompanyPage(client),
		buyPage:         buypage.NewBuyPage(client),
		tradingInfoPage: tradinginfopage.NewTradingInfoPage(),
		currentPage:     messages.LandingPageNumber,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return m, nil
	case messages.PageSwitchMsg:
		m.errorPage.PrevPage = m.currentPage
		m.errorPage.Err = msg.Err

		if m.currentPage != messages.ErrorPageNumber && m.currentPage != messages.CompanyPageNumber && m.currentPage != messages.BuyPageNumber {
			m.companyPage.PrevPage = m.currentPage
		}

		if msg.Symbol != "" {
			m.buyPage.Symbol = msg.Symbol
		}

		m.currentPage = msg.Page
		if msg.Company != nil {
			m.companyPage.CompanyInfo = msg.Company
		}

		model := m.getModelFromPageNumber()
		return m, model.Init()
	case messages.TokenSwitchMsg:
		m.updateToken(msg.Token)
		return m, msg.RetryFunc
	case messages.LoginSuccessMsg:
		m.updateToken(msg.Token)
		m.currentPage = msg.Page
		model := m.getModelFromPageNumber()
		return m, model.Init()
	case messages.ReloadAndSwitchPageMsg:
		m.Reload(msg.Page)
		m.currentPage = msg.Page
		model := m.getModelFromPageNumber()
		return m, model.Init()
	case messages.PageSwitchWithoutInitMsg:
		m.currentPage = msg.Page
		return m, nil
	case messages.ReloadMsg:
		m.Reload(msg.Page)
		return m, nil
	case messages.SmartPageSwitchMsg:
		m.currentPage = msg.Page
		if m.Reloaded(msg.Page) {
			return m, m.getModelFromPageNumber().Init()
		}

		return m, nil
	}

	var cmd tea.Cmd
	var page tea.Model
	switch m.currentPage {
	case messages.LandingPageNumber:
		page, cmd = m.landingPage.Update(msg)
		m.landingPage = page.(landingpage.LandingPage)
	case messages.ErrorPageNumber:
		page, cmd = m.errorPage.Update(msg)
		m.errorPage = page.(errorpage.ErrorPage)
	case messages.WatchlistPageNumber:
		page, cmd = m.watchlistPage.Update(msg)
		m.watchlistPage = page.(watchlistpage.WatchlistPage)
	case messages.LoginPageNumber:
		page, cmd = m.loginPage.Update(msg)
		m.loginPage = page.(loginpage.LoginPage)
	case messages.SearchPageNumber:
		page, cmd = m.searchPage.Update(msg)
		m.searchPage = page.(searchpage.SearchPage)
	case messages.CompanyPageNumber:
		page, cmd = m.companyPage.Update(msg)
		m.companyPage = page.(companypage.CompanyPage)
	case messages.BuyPageNumber:
		page, cmd = m.buyPage.Update(msg)
		m.buyPage = page.(buypage.BuyPage)
	case messages.TradingInfoPageNumber:
		page, cmd = m.tradingInfoPage.Update(msg)
		m.tradingInfoPage = page.(tradinginfopage.TradingInfoPage)
	default:
		m.currentPage = messages.ErrorPageNumber
		m.errorPage.Err = errors.New("Unkown error")
	}

	return m, cmd
}

func (m Model) View() string {
	switch m.currentPage {
	case messages.LandingPageNumber:
		return m.landingPage.View()
	case messages.ErrorPageNumber:
		return m.errorPage.View()
	case messages.WatchlistPageNumber:
		return m.watchlistPage.View()
	case messages.LoginPageNumber:
		return m.loginPage.View()
	case messages.SearchPageNumber:
		return m.searchPage.View()
	case messages.CompanyPageNumber:
		return m.companyPage.View()
	case messages.BuyPageNumber:
		return m.buyPage.View()
	case messages.TradingInfoPageNumber:
		return m.tradingInfoPage.View()
	default:
		return m.errorPage.View()
	}
}

func (m *Model) setSize(width, height int) {
	m.landingPage.BaseModel.Width = width
	m.landingPage.BaseModel.Height = height

	m.errorPage.BaseModel.Width = width
	m.errorPage.BaseModel.Height = height

	m.watchlistPage.BaseModel.Width = width
	m.watchlistPage.BaseModel.Height = height

	m.loginPage.BaseModel.Width = width
	m.loginPage.BaseModel.Height = height

	m.searchPage.BaseModel.Width = width
	m.searchPage.BaseModel.Height = height

	m.companyPage.BaseModel.Width = width
	m.companyPage.BaseModel.Height = height

	m.buyPage.BaseModel.Width = width
	m.buyPage.BaseModel.Height = height

	m.tradingInfoPage.BaseModel.Width = width
	m.tradingInfoPage.BaseModel.Height = height
}

func (m *Model) updateToken(token string) {
	m.watchlistPage.BaseModel.Token = token
	m.errorPage.BaseModel.Token = token
	m.landingPage.BaseModel.Token = token
	m.loginPage.BaseModel.Token = token
	m.searchPage.BaseModel.Token = token
	m.companyPage.BaseModel.Token = token
	m.buyPage.BaseModel.Token = token
	m.tradingInfoPage.BaseModel.Token = token
}

func (m *Model) getModelFromPageNumber() tea.Model {
	switch m.currentPage {
	case messages.LandingPageNumber:
		return m.landingPage
	case messages.WatchlistPageNumber:
		return m.watchlistPage
	case messages.ErrorPageNumber:
		return m.errorPage
	case messages.LoginPageNumber:
		return m.loginPage
	case messages.SearchPageNumber:
		return m.searchPage
	case messages.CompanyPageNumber:
		return m.companyPage
	case messages.BuyPageNumber:
		return m.buyPage
	case messages.TradingInfoPageNumber:
		return m.tradingInfoPage
	default:
		return nil
	}
}

func (m *Model) Reload(page int) {
	switch page {
	case messages.LandingPageNumber:
		m.landingPage.Reload()
	case messages.WatchlistPageNumber:
		m.watchlistPage.Reload()
	case messages.ErrorPageNumber:
		m.errorPage.Reload()
	case messages.LoginPageNumber:
		m.loginPage.Reload()
	case messages.SearchPageNumber:
		m.searchPage.Reload()
	case messages.CompanyPageNumber:
		m.companyPage.Reload()
	case messages.BuyPageNumber:
		m.buyPage.Reload()
	default:
		return
	}
}

func (m *Model) Reloaded(page int) bool {
	switch page {
	case messages.WatchlistPageNumber:
		reloaded := m.watchlistPage.Reloaded
		if reloaded {
			m.watchlistPage.Reloaded = false
		}

		return reloaded
	default:
		return false
	}
}
