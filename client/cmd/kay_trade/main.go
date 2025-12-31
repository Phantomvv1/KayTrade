package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Phantomvv1/KayTrade/internal/model"
	"github.com/Phantomvv1/KayTrade/internal/requests"
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

	refreshToken, err := model.ReadAndDecryptAESGCM([]byte(os.Getenv("ENCRYPTION_KEY")))
	if err != nil {
		log.Println(err)
	}

	requests.MakeRequest(http.MethodPost, "http://localhost:42069/refresh")

	p := tea.NewProgram(model.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Println(err)
		return
	}
}
