package model

import (
	landingpage "github.com/Phantomvv1/KayTrade/internal/landing_page"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	LandingPageNumber = iota
	WatchlistPageNumber
)

type Model struct {
	landingPage landingpage.LandingPage
	currentPage int
}

func NewModel() Model {
	return Model{
		landingPage: landingpage.LandingPage{},
		currentPage: LandingPageNumber,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentPage {
	case LandingPageNumber:
		return m.landingPage.Update(msg)
	default:
		return nil, nil
	}
}

func (m Model) View() string {
	switch m.currentPage {
	case LandingPageNumber:
		return m.landingPage.View()
	default:
		return ""
	}
}
