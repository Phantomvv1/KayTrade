package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"

	basemodel "github.com/Phantomvv1/KayTrade/internal/base_model"
	buypage "github.com/Phantomvv1/KayTrade/internal/buy_page"
	companypage "github.com/Phantomvv1/KayTrade/internal/company_page"
	errorpage "github.com/Phantomvv1/KayTrade/internal/error_page"
	landingpage "github.com/Phantomvv1/KayTrade/internal/landing_page"
	loginpage "github.com/Phantomvv1/KayTrade/internal/login_page"
	"github.com/Phantomvv1/KayTrade/internal/messages"
	orderpage "github.com/Phantomvv1/KayTrade/internal/order_page"
	positionpage "github.com/Phantomvv1/KayTrade/internal/position_page"
	profilepage "github.com/Phantomvv1/KayTrade/internal/profile_page"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	searchpage "github.com/Phantomvv1/KayTrade/internal/search_page"
	sellpage "github.com/Phantomvv1/KayTrade/internal/sell_page"
	signuppage "github.com/Phantomvv1/KayTrade/internal/sign_up_page"
	tradinginfopage "github.com/Phantomvv1/KayTrade/internal/trading_info_page"
	watchlistpage "github.com/Phantomvv1/KayTrade/internal/watchlist_page"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	landingPage     landingpage.LandingPage
	errorPage       errorpage.ErrorPage
	watchlistPage   watchlistpage.WatchlistPage
	loginPage       loginpage.LoginPage
	searchPage      searchpage.SearchPage
	companyPage     companypage.CompanyPage
	buyPage         buypage.BuyPage
	tradingInfoPage tradinginfopage.TradingInfoPage
	profilePage     profilepage.ProfilePage
	sellPage        sellpage.SellPage
	signUpPage      signuppage.SignUpPage
	orderPage       orderpage.OrderPage
	positionPage    positionpage.PositionPage
	client          *http.Client
	tokenStore      *basemodel.TokenStore
	currentPage     int
}

func NewModel() Model {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Println(err)
	}

	client := &http.Client{Jar: jar}
	tokenStore := &basemodel.TokenStore{Token: ""}

	model := Model{
		landingPage:     landingpage.LandingPage{},
		errorPage:       errorpage.ErrorPage{},
		watchlistPage:   watchlistpage.NewWatchlistPage(client, tokenStore),
		loginPage:       loginpage.NewLoginPage(client, tokenStore),
		searchPage:      searchpage.NewSearchPage(client, tokenStore),
		companyPage:     companypage.NewCompanyPage(client, tokenStore),
		buyPage:         buypage.NewBuyPage(client, tokenStore),
		tradingInfoPage: tradinginfopage.NewTradingInfoPage(),
		profilePage:     profilepage.NewProfilePage(client, tokenStore),
		sellPage:        sellpage.NewSellPage(client, tokenStore),
		signUpPage:      signuppage.NewSignUpPage(client, tokenStore),
		orderPage:       orderpage.NewOrderPage(client),
		positionPage:    positionpage.NewPositionPage(client),
		client:          client,
		tokenStore:      tokenStore,
		currentPage:     messages.LandingPageNumber,
	}

	refreshToken, err := readAndDecryptAESGCM([]byte(os.Getenv("ENCRYPTION_KEY")))
	if err != nil {
		log.Println(err)
		model.landingPage.LogIn = true
		return model
	}

	u, err := url.Parse("http://localhost:42069")
	if err != nil {
		model.landingPage.LogIn = true
		return model
	}

	client.Jar.SetCookies(u, []*http.Cookie{{
		Name:  "refresh",
		Value: refreshToken,
		Path:  "/",
	}})

	body, err := requests.MakeRequest(http.MethodPost, "http://localhost:42069/refresh", nil, client, model.tokenStore)
	if err != nil {
		log.Println(err)
		model.landingPage.LogIn = true
		return model
	}

	var info map[string]string
	err = json.Unmarshal(body, &info)
	if err != nil {
		log.Println(err)
		model.landingPage.LogIn = true
		return model
	}

	model.tokenStore.Token = info["token"]
	model.landingPage.LogIn = false

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return m, nil
	case messages.PageSwitchMsg:
		m.errorPage.PrevPage = m.currentPage
		m.errorPage.Err = msg.Err

		if m.currentPage == messages.SearchPageNumber || m.currentPage == messages.WatchlistPageNumber {
			m.companyPage.PrevPage = m.currentPage
		}

		if m.currentPage == messages.SellPageNumber || m.currentPage == messages.BuyPageNumber {
			m.tradingInfoPage.PrevPage = m.currentPage
		}

		if msg.Order != nil {
			m.orderPage.Order = msg.Order
		}

		if msg.Position != nil {
			m.positionPage.Position = msg.Position
		}

		if msg.Symbol != "" {
			m.buyPage.Symbol = msg.Symbol
			m.sellPage.Symbol = msg.Symbol
		}

		if msg.MaxQuantity != 0 {
			m.sellPage.MaxQuantity = msg.MaxQuantity
		}

		m.currentPage = msg.Page
		if msg.Company != nil {
			m.companyPage.CompanyInfo = msg.Company
		}

		model := m.getModelFromPageNumber()
		return m, model.Init()
	case messages.LoginSuccessMsg:
		m.tokenStore.Token = msg.Token
		m.currentPage = msg.Page
		model := m.getModelFromPageNumber()
		return m, model.Init()
	case messages.PageSwitchWithoutInitMsg:
		m.currentPage = msg.Page
		return m, nil
	case messages.ReloadMsg:
		m.Reload(msg.Page)
		return m, nil
	case messages.SmartPageSwitchMsg:
		m.currentPage = msg.Page
		if m.Reloaded(msg.Page) {
			return m, m.getModelFromPageNumber().Init()
		}

		return m, nil
	case messages.QuitMsg:
		if err := m.saveRefreshToken(); err != nil {
			log.Println("Unable to save the refresh token")
			log.Println(err)
			return m, tea.Quit
		}

		return m, tea.Quit
	}

	var cmd tea.Cmd
	var page tea.Model
	switch m.currentPage {
	case messages.LandingPageNumber:
		page, cmd = m.landingPage.Update(msg)
		m.landingPage = page.(landingpage.LandingPage)
	case messages.ErrorPageNumber:
		page, cmd = m.errorPage.Update(msg)
		m.errorPage = page.(errorpage.ErrorPage)
	case messages.WatchlistPageNumber:
		page, cmd = m.watchlistPage.Update(msg)
		m.watchlistPage = page.(watchlistpage.WatchlistPage)
	case messages.LoginPageNumber:
		page, cmd = m.loginPage.Update(msg)
		m.loginPage = page.(loginpage.LoginPage)
	case messages.SearchPageNumber:
		page, cmd = m.searchPage.Update(msg)
		m.searchPage = page.(searchpage.SearchPage)
	case messages.CompanyPageNumber:
		page, cmd = m.companyPage.Update(msg)
		m.companyPage = page.(companypage.CompanyPage)
	case messages.BuyPageNumber:
		page, cmd = m.buyPage.Update(msg)
		m.buyPage = page.(buypage.BuyPage)
	case messages.TradingInfoPageNumber:
		page, cmd = m.tradingInfoPage.Update(msg)
		m.tradingInfoPage = page.(tradinginfopage.TradingInfoPage)
	case messages.ProfilePageNumber:
		page, cmd = m.profilePage.Update(msg)
		m.profilePage = page.(profilepage.ProfilePage)
	case messages.SellPageNumber:
		page, cmd = m.sellPage.Update(msg)
		m.sellPage = page.(sellpage.SellPage)
	case messages.SignUpPageNumber:
		page, cmd = m.signUpPage.Update(msg)
		m.signUpPage = page.(signuppage.SignUpPage)
	case messages.OrderPageNumber:
		page, cmd = m.orderPage.Update(msg)
		m.orderPage = page.(orderpage.OrderPage)
	case messages.PositionPageNumber:
		page, cmd = m.positionPage.Update(msg)
		m.positionPage = page.(positionpage.PositionPage)

	default:
		m.currentPage = messages.ErrorPageNumber
		m.errorPage.Err = errors.New("Unkown error")
	}

	return m, cmd
}

func (m Model) View() string {
	switch m.currentPage {
	case messages.LandingPageNumber:
		return m.landingPage.View()
	case messages.ErrorPageNumber:
		return m.errorPage.View()
	case messages.WatchlistPageNumber:
		return m.watchlistPage.View()
	case messages.LoginPageNumber:
		return m.loginPage.View()
	case messages.SearchPageNumber:
		return m.searchPage.View()
	case messages.CompanyPageNumber:
		return m.companyPage.View()
	case messages.BuyPageNumber:
		return m.buyPage.View()
	case messages.TradingInfoPageNumber:
		return m.tradingInfoPage.View()
	case messages.ProfilePageNumber:
		return m.profilePage.View()
	case messages.SellPageNumber:
		return m.sellPage.View()
	case messages.SignUpPageNumber:
		return m.signUpPage.View()
	case messages.OrderPageNumber:
		return m.orderPage.View()
	case messages.PositionPageNumber:
		return m.positionPage.View()

	default:
		return m.errorPage.View()
	}
}

func (m *Model) setSize(width, height int) {
	m.landingPage.BaseModel.Width = width
	m.landingPage.BaseModel.Height = height

	m.errorPage.BaseModel.Width = width
	m.errorPage.BaseModel.Height = height

	m.watchlistPage.BaseModel.Width = width
	m.watchlistPage.BaseModel.Height = height

	m.loginPage.BaseModel.Width = width
	m.loginPage.BaseModel.Height = height

	m.searchPage.BaseModel.Width = width
	m.searchPage.BaseModel.Height = height

	m.companyPage.BaseModel.Width = width
	m.companyPage.BaseModel.Height = height

	m.buyPage.BaseModel.Width = width
	m.buyPage.BaseModel.Height = height

	m.tradingInfoPage.BaseModel.Width = width
	m.tradingInfoPage.BaseModel.Height = height

	m.profilePage.BaseModel.Width = width
	m.profilePage.BaseModel.Height = height

	m.sellPage.BaseModel.Width = width
	m.sellPage.BaseModel.Height = height

	m.signUpPage.BaseModel.Width = width
	m.signUpPage.BaseModel.Height = height

	m.orderPage.BaseModel.Width = width
	m.orderPage.BaseModel.Height = height

	m.positionPage.BaseModel.Width = width
	m.positionPage.BaseModel.Height = height
}

func (m *Model) getModelFromPageNumber() tea.Model {
	switch m.currentPage {
	case messages.LandingPageNumber:
		return m.landingPage
	case messages.WatchlistPageNumber:
		return m.watchlistPage
	case messages.ErrorPageNumber:
		return m.errorPage
	case messages.LoginPageNumber:
		return m.loginPage
	case messages.SearchPageNumber:
		return m.searchPage
	case messages.CompanyPageNumber:
		return m.companyPage
	case messages.BuyPageNumber:
		return m.buyPage
	case messages.TradingInfoPageNumber:
		return m.tradingInfoPage
	case messages.ProfilePageNumber:
		return m.profilePage
	case messages.SellPageNumber:
		return m.sellPage
	case messages.SignUpPageNumber:
		return m.signUpPage
	case messages.OrderPageNumber:
		return m.orderPage
	case messages.PositionPageNumber:
		return m.positionPage
	default:
		return nil
	}
}

func (m *Model) Reload(page int) {
	switch page {
	case messages.LandingPageNumber:
		m.landingPage.Reload()
	case messages.WatchlistPageNumber:
		m.watchlistPage.Reload()
	case messages.ErrorPageNumber:
		m.errorPage.Reload()
	case messages.LoginPageNumber:
		m.loginPage.Reload()
	case messages.SearchPageNumber:
		m.searchPage.Reload()
	case messages.CompanyPageNumber:
		m.companyPage.Reload()
	case messages.BuyPageNumber:
		m.buyPage.Reload()
	case messages.ProfilePageNumber:
		m.profilePage.Reload()
	case messages.SellPageNumber:
		m.sellPage.Reload()
	default:
		return
	}
}

func (m *Model) Reloaded(page int) bool {
	switch page {
	case messages.WatchlistPageNumber:
		reloaded := m.watchlistPage.Reloaded
		if reloaded {
			m.watchlistPage.Reloaded = false
		}

		return reloaded

	case messages.ProfilePageNumber:
		reloaded := m.profilePage.Reloaded
		if reloaded {
			m.profilePage.Reloaded = false
		}

		return reloaded

	default:
		return false
	}
}

func (m Model) extractRefreshToken() (string, error) {
	u, err := url.Parse("http://localhost:42069")
	if err != nil {
		return "", err
	}

	cookie := m.client.Jar.Cookies(u)[0]
	if cookie.Name != "refresh" {
		return "", errors.New("refresh token not found in cookie jar")
	}

	return cookie.Value, nil
}

func (m Model) saveRefreshToken() error {
	token, err := m.extractRefreshToken()
	if err != nil {
		return err
	}

	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	encrypted, err := encryptAESGCM([]byte(token), key)
	if err != nil {
		return err
	}

	// Config dir for now. Will think abouth which is better: config or config
	config, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(config+"/kaytrade", 0700)
	if err != nil && !os.IsExist(err) {
		log.Println(err)
		return err
	}

	return os.WriteFile(
		filepath.Join(config, "/kaytrade", "/kaytrade"),
		encrypted,
		0600,
	)
}

func encryptAESGCM(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func readAndDecryptAESGCM(key []byte) (string, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	cipherText, err := os.ReadFile(config + "/kaytrade" + "/kaytrade")
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce := cipherText[:nonceSize]
	encrypted := cipherText[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
