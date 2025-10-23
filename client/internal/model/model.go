package model

import (
	"errors"
	"log"

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
	// jar, err := cookiejar.New(nil)
	// if err != nil {
	// 	log.Println(err)
	// }
	//
	// client := http.Client{Jar: jar}

	return Model{
		landingPage:   landingpage.LandingPage{},
		errorPage:     errorpage.ErrorPage{},
		watchlistPage: watchlistpage.NewWatchlistPage(),
		loginPage:     loginpage.NewLoginPage(),
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
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case messages.PageSwitchMsg:
		m.currentPage = msg.Page
		m.errorPage.Err = msg.Err
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
		m.errorPage.Err = errors.New("Unkown error")
		return m.errorPage.View()
	}
}

func (m *Model) SetSize(width, height int) {
	m.landingPage.BaseModel.Width = width
	m.landingPage.BaseModel.Height = height

	m.errorPage.BaseModel.Width = width
	m.errorPage.BaseModel.Height = height

	m.watchlistPage.BaseModel.Width = width
	m.watchlistPage.BaseModel.Height = height

	m.loginPage.BaseModel.Width = width
	m.loginPage.BaseModel.Height = height
}
