package watchlistpage

import (
	"testing"

	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

func TestWatchlist_Update_InitResult(t *testing.T) {
	w := NewWatchlistPage(nil, nil)
	w.BaseModel.Width = 100
	w.BaseModel.Height = 40

	msg := initResult{
		Companies: []messages.CompanyInfo{
			{Symbol: "AAPL", Name: "Apple", Description: "Tech"},
		},
		Gainers: []MarketMover{
			{Symbol: "TSLA"},
		},
		Losers: []MarketMover{},
	}

	model, _ := w.Update(msg)
	w2 := model.(WatchlistPage)

	if !w2.loaded {
		t.Fatalf("expected loaded=true")
	}

	if len(w2.companies.Items()) != 1 {
		t.Fatalf("expected 1 company item")
	}

	if len(w2.movers) != 1 {
		t.Fatalf("expected movers merged")
	}
}

func TestWatchlist_Reload(t *testing.T) {
	w := NewWatchlistPage(nil, nil)

	w.loaded = true
	w.emptyWatchlist = true

	w.Reload()

	if !w.Reloaded {
		t.Fatalf("expected Reloaded=true")
	}

	if w.loaded {
		t.Fatalf("expected loaded=false")
	}

	if w.emptyWatchlist {
		t.Fatalf("expected emptyWatchlist reset")
	}
}

func TestWatchlist_Update_EnterEmptyWatchlist(t *testing.T) {
	w := NewWatchlistPage(nil, nil)
	w.loaded = true
	w.emptyWatchlist = true

	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Fatalf("expected no command when empty watchlist")
	}
}

func TestWatchlist_Update_SearchKey(t *testing.T) {
	w := NewWatchlistPage(nil, nil)

	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})

	msg := cmd().(messages.PageSwitchMsg)

	if msg.Page != messages.SearchPageNumber {
		t.Fatalf("expected search page switch")
	}
}
