package model

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	bankrelationshipcreationpage "github.com/Phantomvv1/KayTrade/client/internal/bank_relationship_creation_page"
	bankrelationshippage "github.com/Phantomvv1/KayTrade/client/internal/bank_relationship_page"
	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
	buypage "github.com/Phantomvv1/KayTrade/client/internal/buy_page"
	companypage "github.com/Phantomvv1/KayTrade/client/internal/company_page"
	documentspage "github.com/Phantomvv1/KayTrade/client/internal/documents_page"
	errorpage "github.com/Phantomvv1/KayTrade/client/internal/error_page"
	landingpage "github.com/Phantomvv1/KayTrade/client/internal/landing_page"
	loginpage "github.com/Phantomvv1/KayTrade/client/internal/login_page"
	"github.com/Phantomvv1/KayTrade/client/internal/messages"
	orderpage "github.com/Phantomvv1/KayTrade/client/internal/order_page"
	positionpage "github.com/Phantomvv1/KayTrade/client/internal/position_page"
	profilepage "github.com/Phantomvv1/KayTrade/client/internal/profile_page"
	"github.com/Phantomvv1/KayTrade/client/internal/requests"
	searchpage "github.com/Phantomvv1/KayTrade/client/internal/search_page"
	sellpage "github.com/Phantomvv1/KayTrade/client/internal/sell_page"
	signuppage "github.com/Phantomvv1/KayTrade/client/internal/sign_up_page"
	tradinginfopage "github.com/Phantomvv1/KayTrade/client/internal/trading_info_page"
	transferspage "github.com/Phantomvv1/KayTrade/client/internal/transfers_page"
	viewtransferspage "github.com/Phantomvv1/KayTrade/client/internal/view_transfers_page"
	watchlistpage "github.com/Phantomvv1/KayTrade/client/internal/watchlist_page"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel() Model {
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: jar,
	}

	tokenStore := &basemodel.TokenStore{}

	return Model{
		landingPage:                  landingpage.LandingPage{},
		errorPage:                    errorpage.ErrorPage{},
		watchlistPage:                watchlistpage.NewWatchlistPage(client, tokenStore),
		loginPage:                    loginpage.NewLoginPage(client, tokenStore),
		searchPage:                   searchpage.NewSearchPage(client, tokenStore),
		companyPage:                  companypage.NewCompanyPage(client, tokenStore),
		buyPage:                      buypage.NewBuyPage(client, tokenStore),
		tradingInfoPage:              tradinginfopage.NewTradingInfoPage(),
		profilePage:                  profilepage.NewProfilePage(client, tokenStore),
		sellPage:                     sellpage.NewSellPage(client, tokenStore),
		signUpPage:                   signuppage.NewSignUpPage(client, tokenStore),
		orderPage:                    orderpage.NewOrderPage(client),
		positionPage:                 positionpage.NewPositionPage(client),
		bankRelationshipPage:         bankrelationshippage.NewBankRelationshipPage(client, tokenStore),
		bankRelationshipCreationPage: bankrelationshipcreationpage.NewBankRelationship(client, tokenStore),
		transfersPage:                transferspage.NewTransfersPage(client, tokenStore),
		viewTransfersPage:            viewtransferspage.New(client, tokenStore),
		documentsPage:                documentspage.New(client, tokenStore),
		client:                       client,
		tokenStore:                   tokenStore,
		currentPage:                  messages.LandingPageNumber,
	}
}

func TestInitReturnsNil(t *testing.T) {
	m := newTestModel()

	if cmd := m.Init(); cmd != nil {
		t.Fatal("expected Init to return nil")
	}
}

func TestSetSize(t *testing.T) {
	m := newTestModel()

	m.setSize(120, 40)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"landing", m.landingPage.BaseModel.Width, m.landingPage.BaseModel.Height},
		{"error", m.errorPage.BaseModel.Width, m.errorPage.BaseModel.Height},
		{"watchlist", m.watchlistPage.BaseModel.Width, m.watchlistPage.BaseModel.Height},
		{"login", m.loginPage.BaseModel.Width, m.loginPage.BaseModel.Height},
		{"search", m.searchPage.BaseModel.Width, m.searchPage.BaseModel.Height},
		{"company", m.companyPage.BaseModel.Width, m.companyPage.BaseModel.Height},
		{"buy", m.buyPage.BaseModel.Width, m.buyPage.BaseModel.Height},
		{"profile", m.profilePage.BaseModel.Width, m.profilePage.BaseModel.Height},
		{"sell", m.sellPage.BaseModel.Width, m.sellPage.BaseModel.Height},
		{"documents", m.documentsPage.BaseModel.Width, m.documentsPage.BaseModel.Height},
	}

	for _, tt := range tests {
		if tt.width != 120 || tt.height != 40 {
			t.Fatalf("%s page size not updated correctly", tt.name)
		}
	}
}

func TestGetModelFromPageNumber(t *testing.T) {
	m := newTestModel()

	tests := []int{
		messages.LandingPageNumber,
		messages.WatchlistPageNumber,
		messages.LoginPageNumber,
		messages.DocumentsPageNumber,
		messages.ErrorPageNumber,
	}

	for _, page := range tests {
		m.currentPage = page

		if got := m.getModelFromPageNumber(); got == nil {
			t.Fatalf("expected model for page %d", page)
		}
	}

	m.currentPage = 999

	if got := m.getModelFromPageNumber(); got != nil {
		t.Fatal("expected nil for invalid page")
	}
}

func TestReloaded(t *testing.T) {
	m := newTestModel()

	m.watchlistPage.Reloaded = true

	if !m.Reloaded(messages.WatchlistPageNumber) {
		t.Fatal("expected watchlist reloaded")
	}

	if m.watchlistPage.Reloaded {
		t.Fatal("expected Reloaded to reset flag")
	}

	if !m.Reloaded(messages.SearchPageNumber) {
		t.Fatal("search page should always return true")
	}

	if m.Reloaded(999) {
		t.Fatal("invalid page should return false")
	}
}

func TestReload(t *testing.T) {
	m := newTestModel()

	m.Reload(messages.WatchlistPageNumber)

	if !m.watchlistPage.Reloaded {
		t.Fatal("expected watchlist page to reload")
	}

	m.Reload(messages.DocumentsPageNumber)

	if !m.documentsPage.Reloaded {
		t.Fatal("expected documents page to reload")
	}
}

func TestExtractRefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := newTestModel()

		u, _ := url.Parse(requests.BaseURL)

		m.client.Jar.SetCookies(u, []*http.Cookie{
			{
				Name:  "refresh",
				Value: "token123",
			},
		})

		token, err := m.extractRefreshToken()
		if err != nil {
			t.Fatal(err)
		}

		if token != "token123" {
			t.Fatalf("expected token123 got %s", token)
		}
	})

	t.Run("no cookies", func(t *testing.T) {
		m := newTestModel()

		_, err := m.extractRefreshToken()
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("wrong cookie name", func(t *testing.T) {
		m := newTestModel()

		u, _ := url.Parse(requests.BaseURL)

		m.client.Jar.SetCookies(u, []*http.Cookie{
			{
				Name:  "session",
				Value: "abc",
			},
		})

		_, err := m.extractRefreshToken()
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUpdateWindowSizeMsg(t *testing.T) {
	m := newTestModel()

	model, _ := m.Update(tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	})

	updated := model.(Model)

	if updated.documentsPage.BaseModel.Width != 100 {
		t.Fatal("width not updated")
	}

	if updated.documentsPage.BaseModel.Height != 50 {
		t.Fatal("height not updated")
	}
}

func TestUpdatePageSwitchMsg(t *testing.T) {
	m := newTestModel()

	order := &messages.Order{
		Symbol: "AAPL",
	}

	position := &messages.Position{
		Symbol: "TSLA",
	}

	company := &messages.CompanyInfo{
		Symbol: "NVDA",
	}

	funding := &messages.FundingInformation{
		TransferType: "deposit",
	}

	model, _ := m.Update(messages.PageSwitchMsg{
		Page:               messages.CompanyPageNumber,
		Err:                errors.New("test"),
		Company:            company,
		Symbol:             "AAPL",
		MaxQuantity:        12,
		Order:              order,
		Position:           position,
		FundingInformation: funding,
	})

	updated := model.(Model)

	if updated.currentPage != messages.CompanyPageNumber {
		t.Fatal("page not switched")
	}

	if updated.companyPage.CompanyInfo.Symbol != "NVDA" {
		t.Fatal("company info not set")
	}

	if updated.buyPage.Symbol != "AAPL" {
		t.Fatal("buy symbol not set")
	}

	if updated.sellPage.MaxQuantity != 12 {
		t.Fatal("max quantity not set")
	}
}

func TestUpdateLoginSuccessMsg(t *testing.T) {
	m := newTestModel()

	model, _ := m.Update(messages.LoginSuccessMsg{
		Token: "jwt-token",
		Page:  messages.ProfilePageNumber,
	})

	updated := model.(Model)

	if updated.tokenStore.Token != "jwt-token" {
		t.Fatal("token not updated")
	}

	if updated.currentPage != messages.ProfilePageNumber {
		t.Fatal("page not updated")
	}
}

func TestUpdateReloadMsg(t *testing.T) {
	m := newTestModel()

	model, _ := m.Update(messages.ReloadMsg{
		Page: messages.DocumentsPageNumber,
	})

	updated := model.(Model)

	if !updated.documentsPage.Reloaded {
		t.Fatal("documents page should reload")
	}
}

func TestUpdateSmartPageSwitchMsg(t *testing.T) {
	m := newTestModel()

	m.documentsPage.Reloaded = true

	model, _ := m.Update(messages.SmartPageSwitchMsg{
		Page: messages.DocumentsPageNumber,
	})

	updated := model.(Model)

	if updated.currentPage != messages.DocumentsPageNumber {
		t.Fatal("page not switched")
	}

	if updated.documentsPage.Reloaded {
		t.Fatal("expected Reloaded to be set to false")
	}
}

func TestUpdateInvalidPageFallsBackToError(t *testing.T) {
	m := newTestModel()

	m.currentPage = 999

	model, _ := m.Update(struct{}{})

	updated := model.(Model)

	if updated.currentPage != messages.ErrorPageNumber {
		t.Fatal("expected error page")
	}

	if updated.errorPage.Err == nil {
		t.Fatal("expected error")
	}
}
