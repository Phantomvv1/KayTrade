package model

import (
	errorpage "github.com/Phantomvv1/KayTrade/internal/error_page"
	landingpage "github.com/Phantomvv1/KayTrade/internal/landing_page"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	landingPage landingpage.LandingPage
	errorPage   errorpage.ErrorPage
	currentPage int
}

func NewModel() Model {
	// jar, err := cookiejar.New(nil)
	// if err != nil {
	// 	log.Println(err)
	// }
	//
	// client := http.Client{Jar: jar}

	return Model{
		landingPage: landingpage.LandingPage{},
		errorPage:   errorpage.ErrorPage{},
		currentPage: messages.LandingPageNumber,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	}

	return m, cmd
}

func (m Model) View() string {
	switch m.currentPage {
	case messages.LandingPageNumber:
		return m.landingPage.View()
	case messages.ErrorPageNumber:
		return m.errorPage.View()
	default:
		return ""
	}
}
