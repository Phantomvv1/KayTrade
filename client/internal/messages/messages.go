package messages

import tea "github.com/charmbracelet/bubbletea"

const (
	LandingPageNumber = iota
	WatchlistPageNumber
	LoginPageNumber
	ErrorPageNumber
)

type PageSwitchMsg struct {
	Page int
	Err  error
}

type TokenSwitchMsg struct {
	Token     string
	RetryFunc func() tea.Msg
}

type LoginSuccessMsg struct {
	Token string
	Page  int
}
