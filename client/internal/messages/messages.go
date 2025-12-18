package messages

import tea "github.com/charmbracelet/bubbletea"

const (
	LandingPageNumber = iota
	WatchlistPageNumber
	LoginPageNumber
	SearchPageNumber
	CompanyPageNumber
	BuyPageNumber
	TradingInfoPageNumber
	ProfilePageNumber
	ErrorPageNumber
)

type PageSwitchMsg struct {
	Page    int
	Err     error
	Company *CompanyInfo
	Symbol  string
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

type SmartPageSwitchMsg struct {
	Page int
}

type CompanyInfo struct {
	Symbol       string  `json:"symbol"`
	OpeningPrice float64 `json:"opening_price,omitempty"`
	ClosingPrice float64 `json:"closing_price,omitempty"`
	Logo         string  `json:"logo"`
	Name         string  `json:"name"`
	History      string  `json:"history"`
	IsNSFW       bool    `json:"isNsfw"`
	Description  string  `json:"description"`
	FoundedYear  int     `json:"founded_year"`
	Domain       string  `json:"domain"`
}
