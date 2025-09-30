package watchlist

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	"github.com/Phantomvv1/KayTrade/internal/requests"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func CreateWatchlistTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists wishlist(user_id uuid references authentication(id), symbol text)")
	return err
}

func CreateWatchlistAlpaca(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodPost, BaseURL+Trading+id+Watchlist, c.Request.Body, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't create a watchlist for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetWatchlistAlpaca(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+Watchlist, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get all the watchlists for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func ManageWatchlistAlpaca(c *gin.Context) {
	id := c.GetString("id")
	watchlistID := c.Param("watchlistId")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+Watchlist+watchlistID, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the watchlist for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func UpdateWatchlistAlpaca(c *gin.Context) {
	id := c.GetString("id")
	watchlistID := c.Param("watchlistId")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodPut, BaseURL+Trading+id+Watchlist+watchlistID, c.Request.Body, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't update the watchlist for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func DeleteWatchlistAlpaca(c *gin.Context) {
	id := c.GetString("id")
	watchlistID := c.Param("watchlistId")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodDelete, BaseURL+Trading+id+Watchlist+watchlistID, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't delete the watchlist")
		return
	}

	c.JSON(http.StatusOK, body)
}

func AddAssetWatchlistAlpaca(c *gin.Context) {
	id := c.GetString("id")
	watchlistID := c.Param("watchlistId")

	headers := BasicAuth()

	errs := map[int]string{
		404: "The requested watchlist is not found, or one of the symbols is not found in the assets",
		422: "Some parameters are not valid",
	}

	body, err := SendRequest[any](http.MethodPost, BaseURL+Trading+id+Watchlist+watchlistID, c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't add an asset to the watchlist")
		return
	}

	c.JSON(http.StatusOK, body)
}

func RemoveSymbolFromWatchlistAlpaca(c *gin.Context) {
	id := c.GetString("id")
	watchlistID := c.Param("watchlistId")
	symbol := c.Param("symbol")

	headers := BasicAuth()

	errs := map[int]string{
		404: "The requested watchlist is not found",
	}

	body, err := SendRequest[any](http.MethodDelete, BaseURL+Trading+id+Watchlist+watchlistID+symbol, c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't remove symbol from the watchlist")
		return
	}

	c.JSON(http.StatusOK, body)
}

func AddSymbolToWatchlist(c *gin.Context) {
	id := c.GetString("id")
	symbol := c.Param("symbol")
	symbol = strings.ToUpper(symbol)

	if symbol == "" {
		ErrorExit(c, http.StatusBadRequest, "incorrectly provided parameters for adding a symbol to the watchlist", nil)
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't conenct to the database", err)
		return
	}
	defer conn.Close(context.Background())

	if err = CreateWatchlistTable(conn); err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't create the table for the watchlist", err)
		return
	}

	_, err = conn.Exec(context.Background(), "insert into wishlist (user_id, symbol) values ($1, $2)", id, symbol)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't insert the information into the database", err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func getSymbols(conn *pgx.Conn, id string) ([]string, error) {
	rows, err := conn.Query(context.Background(), "select symbol from wishlist w where w.user_id = $1", id)
	if err != nil {
		return nil, err
	}

	symbols, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
		symbol := ""
		err := rows.Scan(&symbol)
		if err != nil {
			return "", err
		}

		return symbol, nil
	})
	if err != nil {
		return nil, errors.New("Error reading the symbols from the database")
	}

	return symbols, nil
}

func GetSymbolsFromWatchlist(c *gin.Context) {
	id := c.GetString("id")

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err)
		return
	}
	defer conn.Close(context.Background())

	symbols, err := getSymbols(conn, id)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't get the symbols from the database", err)
		return
	}

	c.JSON(http.StatusOK, symbols)
}

type Response struct {
	Symbol       string         `json:"symbol"`
	OpeningPrice float64        `json:"opening_price"`
	ClosingPrice float64        `json:"closing_price"`
	Logo         map[string]any `json:"logo"`
}

func GetInformationForSymbols(c *gin.Context) {
	id := c.GetString("id")

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", nil)
		return
	}
	defer conn.Close(context.Background())

	symbols, err := getSymbols(conn, id)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't get the symbols from the database", err)
		return
	}

	now := time.Now().UTC()
	start := ""
	checkPassed := false
	if now.Hour() < 13 || now.Hour() >= 20 { // market opens at 13:30 UTC and closes at 20:00 UTC
		if now.Weekday() == time.Monday {
			log.Println(now.AddDate(0, 0, -3).Truncate(time.Hour * 24).Format(time.RFC3339))
			start = now.AddDate(0, 0, -3).Truncate(time.Hour * 24).Format(time.RFC3339)
		} else {
			start = now.AddDate(0, 0, -1).Truncate(time.Hour * 24).Format(time.RFC3339)
		}

		checkPassed = true
	}
	if now.Hour() == 13 && now.Minute() < 30 {
		if now.Weekday() == time.Monday {
			start = now.AddDate(0, 0, -3).Truncate(time.Hour * 24).Format(time.RFC3339)
		} else {
			start = now.AddDate(0, 0, -1).Truncate(time.Hour * 24).Format(time.RFC3339)
		}
		checkPassed = true
	} else if !checkPassed {
		start = time.Now().UTC().Truncate(24 * time.Hour).Format(time.RFC3339)
	}

	res := make(chan result)
	go getInformation(symbols, start, res)
	for _, symbol := range symbols {
		go getLogo(symbol, res)
	}

	var response []Response
	mu := sync.Mutex{}
	needsLock := true
	for range len(symbols) + 1 {
		result := <-res
		if result.result == 0 {
			if needsLock {
				mu.Lock()
			}

			if result.err != nil {
				ErrorExit(c, http.StatusFailedDependency, "couldn't get the logo", result.err) // change that later
				return
			}

			for i, stock := range response {
				if stock.Symbol == result.symbol {
					response[i].Logo = result.logo
				}
			}

			log.Println(result.symbol)
			r := Response{Symbol: result.symbol, Logo: result.logo}
			response = append(response, r)

			if needsLock {
				mu.Unlock()
			}
		} else {
			mu.Lock()
			if result.err != nil {
				ErrorExit(c, http.StatusFailedDependency, "couldn't get the opening and closing price", result.err)
				return
			}

			for symbol, info := range result.information {
				if index := containsSymbol(response, symbol); index != -1 {
					openingPrice := info[0]["o"].(float64)
					closingPrice := info[0]["c"].(float64)
					response[index].OpeningPrice = openingPrice
					response[index].ClosingPrice = closingPrice
				} else {
					response = append(response, Response{OpeningPrice: info[0]["o"].(float64), ClosingPrice: info[0]["c"].(float64)})
				}
			}

			needsLock = false
			mu.Unlock()
		}
	}

	c.JSON(http.StatusOK, gin.H{"information": response})
}

func containsSymbol(response []Response, symbol string) int {
	for i, res := range response {
		if res.Symbol == symbol {
			return i
		}
	}

	return -1
}

type result struct {
	result      byte // 0 - logo; 1 - information
	information map[string][]map[string]any
	logo        map[string]any
	err         error
	symbol      string
}

func getLogo(symbol string, res chan<- result) {
	errs := map[int]string{
		400: "Bad Request",
		401: "Unauthorized",
		404: "Not Found or Invalid Domain Name",
		429: "API key quota exceeded",
	}

	header := map[string]string{
		"Authorization": "Bearer " + os.Getenv("BRANDFETCH_API_KEY"),
	}

	body, err := SendRequest[map[string]any](http.MethodGet, "https://api.brandfetch.io/v2/brands/"+symbol, nil, errs, header)
	if err != nil {
		res <- result{logo: nil, result: 0, symbol: symbol, err: err}
		return
	}

	res <- result{logo: body, result: 0, symbol: symbol, err: nil}
}

func getInformation(symbols []string, start string, res chan<- result) {
	s := strings.Join(symbols, ",")
	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[map[string]map[string][]map[string]any](http.MethodGet, requests.MarketData+"/stocks/bars?timeframe=1D&start="+start+"&symbols="+s, nil, errs, headers)
	if err != nil {
		res <- result{information: nil, result: 1, symbol: "", err: err}
		return
	}

	res <- result{information: body["bars"], result: 1, symbol: "", err: nil}
}

func RemoveSymbolFromWatchlist(c *gin.Context) { // to test
	id := c.GetString("id")
	symbol := c.Param("symbol")

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err)
		return
	}
	defer conn.Close(context.Background())

	check := ""
	err = conn.QueryRow(context.Background(), "delete from watchlist where user_id = $1 and symbol = $2 returning user_id", id, symbol).Scan(&check)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorExit(c, http.StatusConflict, "there is no such symbol in your watchlist", err)
			return
		}

		ErrorExit(c, http.StatusInternalServerError, "couldn't delete the symbol from the database", err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func RemoveAllSymbolsFromWatchlist(c *gin.Context) { // to test
	id := c.GetString("id")

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err)
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "delete from watchlist where user_id = $1", id)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't delete the symbols from the database", err)
		return
	}

	c.JSON(http.StatusOK, nil)
}
