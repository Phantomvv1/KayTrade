package profilepage

import (
	"errors"
	"net/http"
	"testing"

	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func fakeProfilePage() ProfilePage {
	client := &http.Client{}
	token := &basemodel.TokenStore{}

	return NewProfilePage(client, token)
}

func TestProfilePage_ReloadResetsState(t *testing.T) {
	p := fakeProfilePage()

	p.alpacaAccount.AccountType = "paper"
	p.tradingDetails.Status = "ACTIVE"
	p.loading = false
	p.Reloaded = false

	p.orders.SetItems([]list.Item{orderItem{}})
	p.positions.SetItems([]list.Item{positionItem{}})

	p.Reload()

	if p.alpacaAccount.AccountType != "" {
		t.Error("expected alpacaAccount reset")
	}
	if p.tradingDetails.Status != "" {
		t.Error("expected tradingDetails reset")
	}
	if !p.loading {
		t.Error("expected loading=true after reload")
	}
	if !p.Reloaded {
		t.Error("expected Reloaded=true after reload")
	}
	if len(p.orders.Items()) != 0 || len(p.positions.Items()) != 0 {
		t.Error("expected lists to be cleared")
	}
}

func TestProfilePage_ProfileDataMsgError(t *testing.T) {
	p := fakeProfilePage()

	msg := profileDataMsg{
		err: errors.New("network error"),
	}

	_, cmd := p.Update(msg)

	responseMsg := cmd().(messages.PageSwitchMsg)
	if responseMsg.Page != messages.ErrorPageNumber {
		t.Fatal("expected switch to the error page")
	}
}

func TestProfilePage_ProfileDataMsgSuccess(t *testing.T) {
	p := fakeProfilePage()

	msg := profileDataMsg{
		tradingDetails: TradingDetails{Status: "ACTIVE"},
		alpacaAccount: AlpacaAccount{
			AccountType: "paper",
		},
		orders: []messages.Order{
			{Symbol: "AAPL", Quantity: "1"},
		},
		positions: []messages.Position{
			{Symbol: "TSLA", Qty: "2"},
		},
	}

	m, cmd := p.Update(msg)
	model := m.(ProfilePage)

	if cmd != nil {
		t.Error("expected no navigation command on success")
	}

	if len(model.orders.Items()) != 1 {
		t.Error("expected orders inserted")
	}

	if len(model.positions.Items()) != 1 {
		t.Error("expected positions inserted")
	}

	if !model.orders.FilterInput.Focused() {
		t.Error("orders list should be focused by default")
	}
}

func TestProfilePage_SwitchToOrders(t *testing.T) {
	p := fakeProfilePage()

	msg := tea.KeyMsg{Type: tea.KeyCtrlH}

	m, _ := p.Update(msg)
	model := m.(ProfilePage)

	if !model.orders.FilterInput.Focused() {
		t.Error("expected orders to be focused")
	}
}

func TestProfilePage_SwitchToPositions(t *testing.T) {
	p := fakeProfilePage()

	msg := tea.KeyMsg{Type: tea.KeyCtrlL}

	m, _ := p.Update(msg)
	model := m.(ProfilePage)

	if !model.positions.FilterInput.Focused() {
		t.Error("expected positions to be focused")
	}
}

func TestProfilePage_EscWithoutFilterNavigates(t *testing.T) {
	p := fakeProfilePage()

	msg := tea.KeyMsg{Type: tea.KeyEsc}

	_, cmd := p.Update(msg)

	respMsg := cmd().(messages.SmartPageSwitchMsg)
	if respMsg.Page != messages.WatchlistPageNumber {
		t.Fatal("expected swicth to watchlist page")
	}
}

func TestProfilePage_FilterModeActivation(t *testing.T) {
	p := fakeProfilePage()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}

	m, _ := p.Update(msg)
	model := m.(ProfilePage)

	if !model.filtering {
		t.Error("expected filtering mode enabled")
	}
}

func TestProfilePage_OrderPageNavigation(t *testing.T) {
	p := fakeProfilePage()

	p.orders.InsertItem(0, orderItem{
		order: messages.Order{
			Symbol:   "AAPL",
			Quantity: "5",
		},
	})
	p.positions.Select(0)
	p.positions.FilterInput.Focus()

	msg := tea.KeyMsg{Type: tea.KeyEnter}

	_, cmd := p.Update(msg)

	respMsg := cmd().(messages.PageSwitchMsg)
	if respMsg.Page != messages.OrderPageNumber {

	}
}
