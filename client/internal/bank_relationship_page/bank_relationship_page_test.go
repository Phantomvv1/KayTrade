package bankrelationshippage

import (
	"errors"
	"net/http"
	"testing"

	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestPage() BankRelationshipPage {
	return NewBankRelationshipPage(&http.Client{}, &basemodel.TokenStore{})
}

func TestBankRelationshipPage_Init_ReturnsFetchCmd(t *testing.T) {
	p := newTestPage()

	cmd := p.Init()
	if cmd == nil {
		t.Fatal("expected cmd not nil")
	}
}

func TestBankRelationshipPage_InitResult_EmptyDataSetsEmptyState(t *testing.T) {
	p := newTestPage()

	msg := initResult{
		bankRelationships: []bankRelationship{},
		achRelationships:  []achRelationship{},
		err:               nil,
	}

	m, _ := p.Update(msg)
	updated := m.(BankRelationshipPage)

	if !updated.loaded {
		t.Fatal("expected loaded to be true")
	}

	if !updated.emptyBankRels {
		t.Fatal("expected bank relationships to be empty")
	}
}

func TestBankRelationshipPage_InitResult_MergesBankAndACH(t *testing.T) {
	p := newTestPage()

	msg := initResult{
		bankRelationships: []bankRelationship{
			{AccountNumber: "BANK-1", BankCodeType: "wire", Status: "active", ID: "b1"},
		},
		achRelationships: []achRelationship{
			{BankAccountNumber: "ACH-1", BankAccountType: "checking", Status: "active", ID: "a1"},
		},
	}

	m, _ := p.Update(msg)
	updated := m.(BankRelationshipPage)

	if !updated.loaded {
		t.Fatal("expected loaded to be true")
	}

	if updated.emptyBankRels {
		t.Fatal("expected bank relationships to not be empty")
	}

	if len(updated.bankRelationships.Items()) != 2 {
		t.Fatal("expected there to be 2 relationships")
	}
}

func TestBankRelationshipPage_InitResult_ErrorSwitchesPage(t *testing.T) {
	p := newTestPage()

	msg := initResult{
		err: errors.New("error"),
	}

	_, cmd := p.Update(msg)

	respMsg := cmd().(messages.PageSwitchMsg)
	if respMsg.Page != messages.ErrorPageNumber {
		t.Fatal("expected page switch to error page")
	}
}

func TestBankRelationshipPage_QuitKey(t *testing.T) {
	p := newTestPage()

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	_ = cmd().(messages.QuitMsg)
}

func TestBankRelationshipPage_EscNavigatesWhenNoFilter(t *testing.T) {
	p := newTestPage()

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc})

	msg := cmd().(messages.SmartPageSwitchMsg)
	if msg.Page != messages.WatchlistPageNumber {
		t.Fatal("expected switch to watchlist page")
	}
}

func TestBankRelationshipPage_CreateRelationshipKey(t *testing.T) {
	p := newTestPage()

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	msg := cmd().(messages.PageSwitchMsg)
	if msg.Page != messages.BankRelationshipCreationPageNumber {
		t.Fatal("expected switch to bank relationship creation page")
	}
}

func TestBankRelationshipPage_ViewTransfersKey(t *testing.T) {
	p := newTestPage()

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})

	msg := cmd().(messages.SmartPageSwitchMsg)
	if msg.Page != messages.ViewTransfersPageNumber {
		t.Fatal("expected switch to view transfers page")
	}
}

func TestBankRelationshipPage_NewTransfer_BankFlow(t *testing.T) {
	p := newTestPage()

	p.bankRelationships.SetItems([]list.Item{
		bankRelationshipItem{
			bankRelationship: &bankRelationship{
				ID:            "bank123",
				AccountNumber: "12345",
			},
		},
	})

	p.bankRelationships.Select(0)

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})

	msg := cmd().(messages.PageSwitchMsg)
	if msg.Page != messages.TransfersPageNumber {
		t.Fatal("expected switch to transfers page")
	}
}

func TestBankRelationshipPage_Reload(t *testing.T) {
	p := newTestPage()

	p.loaded = true
	p.emptyBankRels = true
	p.Reload()

	if p.loaded {
		t.Fatal("expected loaded to be false")
	}

	if p.emptyBankRels {
		t.Fatal("expected emptyBankRels to be false")
	}

	if !p.Reloaded {
		t.Fatal("expected Reloaded to be ture")
	}
}

func TestBankRelationshipPage_DefaultUpdateDelegatesToList(t *testing.T) {
	p := newTestPage()

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if cmd != nil {
		t.Fatal("expected cmd to be nil")
	}
}
