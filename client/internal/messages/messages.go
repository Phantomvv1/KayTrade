package messages

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
