package model

import (
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"

	errorpage "github.com/Phantomvv1/KayTrade/internal/error_page"
	landingpage "github.com/Phantomvv1/KayTrade/internal/landing_page"
	loginpage "github.com/Phantomvv1/KayTrade/internal/login_page"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	watchlistpage "github.com/Phantomvv1/KayTrade/internal/watchlist_page"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	landingPage   landingpage.LandingPage
	errorPage     errorpage.ErrorPage
	watchlistPage watchlistpage.WatchlistPage
	loginPage     loginpage.LoginPage
	currentPage   int
}

func NewModel() Model {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Println(err)
	}

	client := &http.Client{Jar: jar}

	return Model{
		landingPage:   landingpage.LandingPage{},
		errorPage:     errorpage.ErrorPage{},
		watchlistPage: watchlistpage.NewWatchlistPage(client),
		loginPage:     loginpage.NewLoginPage(client),
		currentPage:   messages.LandingPageNumber,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Printf("Size msg!")
		m.setSize(msg.Width, msg.Height)
		return m, nil
	case messages.PageSwitchMsg:
		m.errorPage.PrevPage = m.currentPage
		m.errorPage.Err = msg.Err
		m.currentPage = msg.Page
		return m, nil
	case messages.TokenSwitchMsg:
		m.updateToken(msg.Token)
		return m, msg.RetryFunc
	case messages.LoginSuccessMsg:
		m.updateToken(msg.Token)
		m.currentPage = msg.Page
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
}

func (m *Model) updateToken(token string) {
	m.watchlistPage.BaseModel.Token = token
	m.errorPage.BaseModel.Token = token
	m.landingPage.BaseModel.Token = token
	m.loginPage.BaseModel.Token = token
}

func (m Model) getModelFromPageNumber() tea.Model {
	switch m.currentPage {
	case messages.LandingPageNumber:
		return m.landingPage
	case messages.WatchlistPageNumber:
		return m.watchlistPage
	case messages.ErrorPageNumber:
		return m.errorPage
	default:
		return nil
	}
}
