package main

import (
	"log"
	"os"

	"github.com/Phantomvv1/KayTrade/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	file, err := os.OpenFile("../../logs/logs.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	log.SetOutput(file)

	p := tea.NewProgram(model.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Println(err)
		return
	}
}
