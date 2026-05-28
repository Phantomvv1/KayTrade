package sellpage

import (
	"net/http"
	"testing"

	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

func newSellPage() SellPage {
	client := &http.Client{}
	token := &basemodel.TokenStore{}
	return NewSellPage(client, token)
}

func TestSellPage_Init(t *testing.T) {
	s := newSellPage()

	cmd := s.Init()
	if cmd == nil {
		t.Fatal("expected Init command")
	}
}

func TestSellPage_Reload(t *testing.T) {
	s := newSellPage()

	s.quantity.SetValue("999")
	s.purchaseTypeIdx = 3
	s.timeInForceIdx = 2
	s.cursor = 2
	s.err = "error"
	s.success = "ok"

	s.Reload()

	if s.quantity.Value() != "1" {
		t.Error("expected quantity reset to 1")
	}
	if s.purchaseTypeIdx != 0 {
		t.Error("expected purchaseTypeIdx reset")
	}
	if s.timeInForceIdx != 0 {
		t.Error("expected timeInForceIdx reset")
	}
	if s.cursor != 0 {
		t.Error("expected cursor reset")
	}
	if s.err != "" || s.success != "" {
		t.Error("expected messages cleared")
	}
}

func TestSellPage_CursorWrapUp(t *testing.T) {
	s := newSellPage()
	s.totalFields = 3

	s.cursor = 2
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := m.(SellPage)

	if model.cursor != 0 {
		t.Errorf("expected wrap to 0, got %d", s.cursor)
	}
}

func TestSellPage_CursorWrapDown(t *testing.T) {
	s := newSellPage()
	s.totalFields = 3

	s.cursor = 0
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyUp})
	model := m.(SellPage)

	if model.cursor != 2 {
		t.Errorf("expected wrap to 2, got %d", s.cursor)
	}
}

func TestSellPage_PurchaseTypeSlider(t *testing.T) {
	s := newSellPage()

	s.cursor = 1

	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRight})
	model := m.(SellPage)
	if model.purchaseTypeIdx != 1 {
		t.Error("expected purchaseTypeIdx increment")
	}

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = m.(SellPage)
	if model.purchaseTypeIdx != 0 {
		t.Error("expected purchaseTypeIdx decrement")
	}
}

func TestSellPage_TimeInForceSlider(t *testing.T) {
	s := newSellPage()

	s.cursor = 2

	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRight})
	model := m.(SellPage)
	if model.timeInForceIdx != 1 {
		t.Error("expected timeInForceIdx increment")
	}

	m, _ = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = m.(SellPage)
	if model.timeInForceIdx != 0 {
		t.Error("expected timeInForceIdx decrement")
	}
}

func TestSellPage_InvalidKeyIgnored(t *testing.T) {
	s := newSellPage()

	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model := m.(SellPage)

	if s.cursor != model.cursor {
		t.Error("cursor should not change on invalid key")
	}
}

func TestSellPage_EscNavigation(t *testing.T) {
	s := newSellPage()

	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEsc})

	msg := cmd().(messages.SmartPageSwitchMsg)
	if msg.Page != messages.ProfilePageNumber {
		t.Fatal("expected swicth to the profile page")
	}
}

func TestSellPage_WatchlistNavigation(t *testing.T) {
	s := newSellPage()

	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})

	msg := cmd().(messages.SmartPageSwitchMsg)
	if msg.Page != messages.WatchlistPageNumber {
		t.Fatal("expected switch to the watchlist page")
	}
}

func TestSellPage_InfoNavigation(t *testing.T) {
	s := newSellPage()

	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	msg := cmd().(messages.PageSwitchMsg)
	if msg.Page != messages.TradingInfoPageNumber {
		t.Fatal("expected switch to the trading info page")
	}
}

func TestSellPage_SubmitEmptyQuantity(t *testing.T) {
	s := newSellPage()

	s.quantity.SetValue("")
	err := s.submitOrder()

	if err == nil {
		t.Error("expected error for empty quantity")
	}
}

func TestSellPage_SubmitInvalidFloat(t *testing.T) {
	s := newSellPage()

	s.quantity.SetValue("1.2.3")
	err := s.submitOrder()

	if err == nil {
		t.Error("expected invalid number error")
	}
}

func TestSellPage_SubmitOverflowMarket(t *testing.T) {
	s := newSellPage()

	s.MaxQuantity = 5
	s.purchaseTypeIdx = 0 // market
	s.timeInForceIdx = 0  // day

	s.quantity.SetValue("10")

	err := s.submitOrder()
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestSellPage_SubmitOverflowLimitStyle(t *testing.T) {
	s := newSellPage()

	s.MaxQuantity = 5
	s.purchaseTypeIdx = 1 // limit

	s.quantity.SetValue("10")

	err := s.submitOrder()
	if err == nil {
		t.Error("expected overflow error")
	}
}
