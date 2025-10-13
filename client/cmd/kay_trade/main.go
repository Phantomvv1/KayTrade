package main

import (
	"log"

	"github.com/Phantomvv1/KayTrade/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Println(err)
		return
	}
}
