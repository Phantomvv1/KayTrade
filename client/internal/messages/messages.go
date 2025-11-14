package messages

import tea "github.com/charmbracelet/bubbletea"

const (
	LandingPageNumber = iota
	WatchlistPageNumber
	LoginPageNumber
	SearchPageNumber
	CompanyPageNumber
	ErrorPageNumber
)

type PageSwitchMsg struct {
	Page int
	Err  error
	Comp Company
}

type TokenSwitchMsg struct {
	Token     string
	RetryFunc func() tea.Msg
}

type LoginSuccessMsg struct {
	Token string
	Page  int
}

type ReloadAndSwitchPageMsg struct {
	Page int
}

type PageSwitchWithoutInitMsg struct {
	Page int
}

type ReloadMsg struct {
	Page int
}

type Company interface {
	SymbolInfo() string
	OpeningPriceInfo() float64
	ClosingPriceInfo() float64
	LogoInfo() string
	NameInfo() string
	HistoryInfo() string
	IsNSFWInfo() bool
	DescriptionInfo() string
	FoundedYearInfo() int
	DomainInfo() string
}
