package requests

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	// BaseTradingURL = "https://paper-api.alpaca.markets"
	BaseURL          = "https://broker-api.sandbox.alpaca.markets/v1/"
	MarketData       = "https://data.sandbox.alpaca.markets/v2"
	RealTimeData     = "wss://stream.data.sandbox.alpaca.markets/v2/iex"
	Accounts         = "accounts/"
	Documents        = "documents/"        // Accounts + ":accountId" + Documents
	Trading          = "trading/accounts/" // :accountId
	Assets           = "assets/"
	Calendar         = "calendar/"
	Events           = "events/"
	Transfers        = "transfers/"
	InstantFunding   = "instant_funding/"
	OAuth            = "oauth/"
	Clock            = "clock/"
	Journals         = "journals/"
	CorporateActions = "corporate_actions"
	Watchlist        = "watchlists/" // Trading + :accountId + Watchlist
	Rebalancing      = "rebalancing/"
	Reporting        = "reporting/eod"
	CashInterest     = "cash_interest/apr_tiers" //1 endpoint
	CountryInfo      = "country-info"
	Crypto           = "wallets/" // Accounts + :accountId + Crypto
)

func SendRequest[T any](method, url string, body io.Reader, errs map[int]string, headers map[string]string) (T, error) {
	var zero T
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return zero, err
	}

	for header, value := range headers {
		req.Header.Add(header, value)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return zero, err
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return zero, err
		}

		var errMap map[string]any
		err = json.Unmarshal(resBody, &errMap)
		if err != nil {
			return zero, err
		}

		if errMsg := errs[res.StatusCode]; errMsg != "" {
			log.Println(errMsg)
		}

		return zero, errors.New(errMap["message"].(string))
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return zero, err
	}

	h := res.Header["Content-Type"][0]
	h, _, _ = strings.Cut(h, ";")
	if h == "text/plain" {
		log.Println(string(resBody))
		return zero, errors.New("Unkown error")
	}

	var resJson T
	if len(resBody) > 0 {
		err = json.Unmarshal(resBody, &resJson)
		if err != nil {
			return zero, err
		}
	}

	if res.StatusCode/100 != 2 {
		return resJson, errors.New("Unkown error")
	}

	return resJson, nil
}

func BasicAuth() map[string]string {
	credentials := os.Getenv("API_KEY") + ":" + os.Getenv("SECRET_KEY")
	out := base64.StdEncoding.EncodeToString([]byte(credentials))
	m := map[string]string{"Authorization": "Basic " + out}
	return m
}
