package watchlist

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func CreateWatchlistTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists wishlist(user_id uuid references authentication(id), symbol text)")
	return err
}

func CreateWatchlistAlpaca(c *gin.Context) {
	id := c.Param("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodPost, BaseURL+Trading+id+Watchlist, c.Request.Body, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't create a watchlist for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetWatchlistAlpaca(c *gin.Context) {
	id := c.Param("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+Watchlist, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get all the watchlists for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func ManageWatchlistAlpaca(c *gin.Context) {
	id := c.Param("id")
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
	id := c.Param("id")
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
	id := c.Param("id")
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
	id := c.Param("id")
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
	id := c.Param("id")
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
	id := c.Param("id")
	symbol := c.Param("symbol")

	if id == "" || symbol == "" {
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
	id := c.Param("id")

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

func GetInformationForSymbols(c *gin.Context) {
	id := c.Param("id")

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

	res := make(chan result)
	for _, symbol := range symbols {
		go fetchLogo(symbol, res)
	}

	for range len(symbols) * 2 {

	}
}

type result struct {
	logo []byte
	err  error
}

func fetchLogo(symbol string, res chan<- result) {
	req, err := http.NewRequest(http.MethodGet, "https://broker-api.sandbox.alpaca.markets/v1beta1/logos/"+symbol, nil)
	if err != nil {
		res <- result{logo: nil, err: err}
		return
	}

	credentials := os.Getenv("API_KEY") + ":" + os.Getenv("SECRET_KEY")
	out := base64.StdEncoding.EncodeToString([]byte(credentials))

	req.Header.Add("Authorization", "Basic "+out)
	req.Header.Add("accept", "image/png")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		res <- result{logo: nil, err: err}
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		res <- result{logo: nil, err: err}
		return
	}

	res <- result{logo: body, err: nil}
}
