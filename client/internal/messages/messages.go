package messages

const (
	LandingPageNumber = iota
	WatchlistPageNumber
	ErrorPageNumber
)

type PageSwitchMsg struct {
	Page int
	Err  error
}
