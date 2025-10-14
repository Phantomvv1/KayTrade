package model

import (
	"log"

	errorpage "github.com/Phantomvv1/KayTrade/internal/error_page"
	landingpage "github.com/Phantomvv1/KayTrade/internal/landing_page"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	landingPage landingpage.LandingPage
	errorPage   errorpage.ErrorPage
	currentPage int
	width       int
	height      int
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
	case tea.WindowSizeMsg:
		log.Printf("Size msg!")
		m.width, m.height = msg.Width, msg.Height
	case messages.PageSwitchMsg:
		m.currentPage = msg.Page
		m.errorPage.Err = msg.Err
		return m, nil
	}

	var cmd tea.Cmd
	var page tea.Model
	switch m.currentPage {
	case messages.LandingPageNumber:
		m.landingPage.BaseModel.Width = m.width
		m.landingPage.BaseModel.Height = m.height

		page, cmd = m.landingPage.Update(msg)
		m.landingPage = page.(landingpage.LandingPage)
	case messages.ErrorPageNumber:
		m.errorPage.BaseModel.Width = m.width
		m.errorPage.BaseModel.Height = m.height
		log.Printf("%d, %d. Error page before going in.", m.errorPage.BaseModel.Width, m.errorPage.BaseModel.Height)

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
