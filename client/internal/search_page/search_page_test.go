package searchpage

import (
	"testing"
	"time"

	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestSearchPage() SearchPage {
	return NewSearchPage(nil, nil)
}

func TestSearchPage_PageSwitchMsg_Passthrough(t *testing.T) {
	s := newTestSearchPage()

	in := messages.PageSwitchMsg{Page: messages.CompanyPageNumber}

	_, cmd := s.Update(in)

	msg := cmd().(messages.PageSwitchMsg)
	if msg.Page != messages.CompanyPageNumber {
		t.Fatal("expected switch to company page")
	}
}

func TestSearchPage_Key_Esc_GoesToWatchlist(t *testing.T) {
	s := newTestSearchPage()

	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("esc")})

	msg := cmd().(messages.SmartPageSwitchMsg)
	if msg.Page != messages.WatchlistPageNumber {
		t.Fatal("expected switch to watchlist page")
	}
}

func TestSearchPage_Key_Tab_TogglesPlaceholder(t *testing.T) {
	s := newTestSearchPage()

	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("tab")})
	model := m.(SearchPage)

	if s.name == model.name {
		t.Fatal("expected name mode toggle")
	}

	if s.searchField.Placeholder == model.searchField.Placeholder {
		t.Fatal("expected placeholder change")
	}
}

func TestSearchPage_Key_Navigation_NoPanic(t *testing.T) {
	s := newTestSearchPage()

	keys := []string{"up", "down", "ctrl+j", "ctrl+k"}

	for _, k := range keys {
		_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	}
}

func TestSearchPage_Update_ItemMsg_SetsItems(t *testing.T) {
	s := newTestSearchPage()

	items := []list.Item{asset{asset: Asset{Symbol: "AAPL", Name: "Apple"}}}

	msg := itemMsg{items: items}

	m, _ := s.Update(msg)
	model := m.(SearchPage)

	if len(model.suggestions.Items()) != 1 || len(model.suggestions.Items()) == len(s.suggestions.Items()) {
		t.Fatal("expected suggestions to be replaced")
	}
}

func TestSearchPage_Reload_ClearsState(t *testing.T) {
	s := newTestSearchPage()

	s.searchField.SetValue("TEST")
	s.name = true

	s.Reload()

	if s.searchField.Value() != "" {
		t.Fatal("expected search field cleared")
	}
	if s.name {
		t.Fatal("expected name reset")
	}
}

func TestSearchPage_TimeTick_NoCrash(t *testing.T) {
	s := newTestSearchPage()

	_, _ = s.Update(time.Now())
}

func TestSearchPage_SearchField_UpdatesFlag(t *testing.T) {
	s := newTestSearchPage()

	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")})
	model := m.(SearchPage)

	if s.searchUpdate == model.searchUpdate {
		t.Fatal("expected searchUpdate to change on input")
	}
}
