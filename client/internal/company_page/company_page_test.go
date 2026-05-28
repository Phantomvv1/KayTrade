package companypage

import (
	"errors"
	"testing"

	"github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestPage() CompanyPage {
	return CompanyPage{
		BaseModel: basemodel.BaseModel{},
		tabs:      []int{tabOverview, tabHistory, tabPrice, tabChart},
	}
}

func TestCompanyPage_Reload_ResetsState(t *testing.T) {
	p := newTestPage()

	p.CompanyInfo = &messages.CompanyInfo{
		Symbol: "AAPL",
	}

	p.chartLoading = true
	p.liveConnected = true
	p.chartError = "err"
	p.liveError = "err"
	p.liveData = []timeserieslinechart.TimePoint{{}}

	p.Reload()

	if p.CompanyInfo != nil {
		t.Fatal("expected CompanyInfo to be nil after reload")
	}
	if len(p.liveData) != 0 {
		t.Fatal("expected liveData to be cleared")
	}
	if p.liveConnected {
		t.Fatal("expected liveConnected false")
	}
	if p.chartLoading {
		t.Fatal("expected chartLoading false")
	}
}

func TestCompanyPage_Update_fetchDataMsg_Success(t *testing.T) {
	p := newTestPage()

	msg := fetchDataMsg{
		data: []BarData{
			{Close: 100},
		},
		err: nil,
	}

	m, _ := p.Update(msg)
	cp := m.(CompanyPage)

	if cp.chartLoading {
		t.Fatal("expected chartLoading to be false after success")
	}
	if len(cp.chartData) != 1 {
		t.Fatalf("expected chartData length 1, got %d", len(cp.chartData))
	}
}

func TestCompanyPage_Update_fetchDataMsg_Error(t *testing.T) {
	p := newTestPage()

	msg := fetchDataMsg{
		err: errors.New("fetch failed"),
	}

	_, cmd := p.Update(msg)

	if cmd == nil {
		t.Fatal("expected switch to error page")
	}
}

func TestCompanyPage_Update_wsConnectedMsg(t *testing.T) {
	p := newTestPage()

	m, _ := p.Update(wsConnectedMsg{})

	cp := m.(CompanyPage)

	if !cp.liveConnected {
		t.Fatal("expected liveConnected true")
	}
}

func TestCompanyPage_Update_wsErrorMsg(t *testing.T) {
	p := newTestPage()

	m, _ := p.Update(wsErrorMsg{err: errors.New("ws down")})
	cp := m.(CompanyPage)

	if cp.liveConnected {
		t.Fatal("expected liveConnected false")
	}
	if cp.liveError == "" {
		t.Fatal("expected liveError to be set")
	}
}

func TestCompanyPage_Update_addCompanyMsg_Error(t *testing.T) {
	p := newTestPage()

	_, cmd := p.Update(addCompanyMsg{err: errors.New("fail")})

	returnedCmd := cmd().(messages.PageSwitchMsg)
	if returnedCmd.Page != messages.ErrorPageNumber {
		t.Fatal("expected switch to error page")
	}
}

func TestCompanyPage_Update_addCompanyMsg_Success(t *testing.T) {
	p := newTestPage()

	_, cmd := p.Update(addCompanyMsg{err: nil})

	returnedCmd := cmd().(messages.ReloadMsg)
	if returnedCmd.Page != messages.WatchlistPageNumber {
		t.Fatal("expected to reload the watchlist page")
	}
}

func TestCompanyPage_Update_KeySwitchBuyPage(t *testing.T) {
	p := newTestPage()
	p.CompanyInfo = &messages.CompanyInfo{Symbol: "AAPL"}

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})

	msg := cmd().(messages.PageSwitchMsg)
	if msg.Page != messages.BuyPageNumber {
		t.Fatal("expected switch to buy page")
	}
}

func TestCompanyPage_TabSwitching_DoesNotPanic(t *testing.T) {
	p := newTestPage()
	p.CompanyInfo = &messages.CompanyInfo{Symbol: "AAPL"}

	keys := []string{"l", "h", "right", "left"}

	for _, k := range keys {
		_, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	}
}
