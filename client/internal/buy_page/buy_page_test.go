package buypage

import (
	"testing"

	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

func TestBuyPage_CursorNavigationWraps(t *testing.T) {
	b := NewBuyPage(nil, nil)

	b.totalFields = b.calculateTotalFields()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	b.cursor = 0

	m, _ := b.Update(msg)
	p := m.(BuyPage)

	if p.cursor != p.totalFields-1 {
		t.Fatalf("expected wrap to %d, got %d", p.totalFields-1, p.cursor)
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	m, _ = p.Update(msg)
	p2 := m.(BuyPage)

	if p2.cursor != 0 {
		t.Fatalf("expected wrap to 0, got %d", p2.cursor)
	}
}

func TestBuyPage_QuitMessage(t *testing.T) {
	b := NewBuyPage(nil, nil)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := b.Update(msg)

	if cmd == nil {
		t.Fatal("expected quit command")
	}

	out := cmd()
	if _, ok := out.(messages.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg, got %T", out)
	}
}

func TestBuyPage_EscSwitchesToCompanyPage(t *testing.T) {
	b := NewBuyPage(nil, nil)

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	_, cmd := b.Update(msg)

	if cmd == nil {
		t.Fatal("expected cmd")
	}

	out := cmd()
	pm, ok := out.(messages.PageSwitchMsg)
	if !ok {
		t.Fatalf("expected PageSwitchMsg, got %T", out)
	}

	if pm.Page != messages.CompanyPageNumber {
		t.Fatalf("expected CompanyPageNumber, got %d", pm.Page)
	}
}

func TestBuyPage_WatchlistSmartSwitch(t *testing.T) {
	b := NewBuyPage(nil, nil)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")}
	_, cmd := b.Update(msg)

	out := cmd()
	sm, ok := out.(messages.SmartPageSwitchMsg)
	if !ok {
		t.Fatalf("expected SmartPageSwitchMsg, got %T", out)
	}

	if sm.Page != messages.WatchlistPageNumber {
		t.Fatalf("expected watchlist page")
	}
}

func TestBuyPage_TradingInfoSwitch(t *testing.T) {
	b := NewBuyPage(nil, nil)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")}
	_, cmd := b.Update(msg)

	out := cmd()
	pm, ok := out.(messages.PageSwitchMsg)
	if !ok {
		t.Fatalf("expected PageSwitchMsg, got %T", out)
	}

	if pm.Page != messages.TradingInfoPageNumber {
		t.Fatalf("expected trading info page")
	}
}

func TestBuyPage_SliderLeftRightPurchaseType(t *testing.T) {
	b := NewBuyPage(nil, nil)
	b.cursor = 1

	initial := b.purchaseTypeIdx

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	m, _ := b.Update(msg)
	model := m.(BuyPage)

	if model.purchaseTypeIdx != (initial+1)%len(model.purchaseType) {
		t.Fatalf("expected purchaseType increment")
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	m, _ = model.Update(msg)
	model = m.(BuyPage)

	if model.purchaseTypeIdx != initial {
		t.Fatalf("expected purchaseType decrement")
	}
}

func TestBuyPage_SliderLeftRightTimeInForce(t *testing.T) {
	b := NewBuyPage(nil, nil)
	b.cursor = 2

	initial := b.timeInForceIdx

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	m, _ := b.Update(msg)
	model := m.(BuyPage)

	if model.timeInForceIdx != (initial+1)%len(model.timeInForce) {
		t.Fatalf("expected timeInForce increment")
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	m, _ = model.Update(msg)
	model = m.(BuyPage)

	if model.timeInForceIdx != initial {
		t.Fatalf("expected timeInForce decrement")
	}
}

func TestBuyPage_GetFieldIndexMapping(t *testing.T) {
	b := NewBuyPage(nil, nil)

	tests := []struct {
		cursor int
		nonNil bool
	}{
		{0, true},
		{1, false},
		{2, false},
	}

	for _, tt := range tests {
		b.cursor = tt.cursor
		idx := b.getFieldIndex()

		if tt.nonNil && idx == -1 {
			t.Fatalf("expected valid index for cursor %d", tt.cursor)
		}
	}
}

func TestBuyPage_InputRoutingQuantity(t *testing.T) {
	b := NewBuyPage(nil, nil)

	b.cursor = 0
	idx := b.getFieldIndex()

	in := b.getInputAtIndex(idx)
	if in == nil {
		t.Fatal("expected quantity input")
	}

	in.SetValue("123")
	b.setInputAtIndex(idx, *in)

	if b.quantity.Value() != "123" {
		t.Fatalf("expected quantity updated")
	}
}

func TestBuyPage_ReloadResetsState(t *testing.T) {
	b := NewBuyPage(nil, nil)

	b.cursor = 5
	b.quantity.SetValue("999")
	b.err = "error"
	b.success = "ok"
	b.purchaseTypeIdx = 2
	b.timeInForceIdx = 3

	b.Reload()

	if b.cursor != 0 {
		t.Fatal("expected cursor reset")
	}
	if b.quantity.Value() != "1" {
		t.Fatal("expected quantity reset to 1")
	}
	if b.err != "" || b.success != "" {
		t.Fatal("expected messages cleared")
	}
	if b.purchaseTypeIdx != 0 || b.timeInForceIdx != 0 {
		t.Fatal("expected sliders reset")
	}
}
