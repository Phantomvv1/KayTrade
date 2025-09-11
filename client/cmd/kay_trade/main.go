package main

import (
	"log"

	landingpage "github.com/Phantomvv1/KayTrade/internal/landing_page"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(landingpage.NewLandingPageModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Println(err)
		return
	}
}
