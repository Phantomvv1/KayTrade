package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Phantomvv1/KayTrade/client/internal/model"
	"github.com/Phantomvv1/KayTrade/client/internal/requests"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	envSystem = "system"
	envDocker = "docker"
	envDev    = "dev"
	version   = "0.1.6"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println("KayTrade version " + version)
		return
	}

	env := os.Getenv("KAYTRADE_ENV")
	if env != envDocker && env != envDev {
		env = envSystem
	}

	setupBaseUrl(env)

	err := makeNeededDirs(env)
	if err != nil {
		log.Println(err)
		return
	}

	file, err := setupLogger(env)
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

func setupLogger(env string) (*os.File, error) {
	switch env {
	case envSystem:
		config, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}

		logPath := filepath.Join(config, "kaytrade", "logs", "logs.log")

		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}

		return file, nil

	case envDocker:
		file, err := os.OpenFile("logs.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}

		return file, nil

	case envDev:
		file, err := os.OpenFile("../../logs/logs.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}

		return file, nil
	}

	return nil, errors.New("Env not recognized!")
}

func makeNeededDirs(env string) error {
	config, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	switch env {
	case envSystem:
		err := os.MkdirAll(config+"/kaytrade/logs", 0700) // creates both .config/kaytrade and .config/kaytrade/logs
		if err != nil {
			return err
		}

		return nil

	case envDev:
		err := os.Mkdir("../../logs", 0700)
		if err != nil && !os.IsExist(err) {
			return err
		}

		return nil

	case envDocker:
		return nil
	}

	return errors.New("Env not recognized!")
}

func setupBaseUrl(env string) {
	switch env {
	case envDev:
		return // it already is set for development
	case envSystem, envDocker:
		requests.BaseURL = "http://34.159.60.213:42069"
	}
}
