package trading

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateOrdersTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists orders(id uuid primary key, user_id uuid references authentication(id) on delete cascade, "+
		"symbol text, side text, created_at timestamp, updated_at timestamp)")
	return err
}

func CreateOrder(c *gin.Context) {
	id := c.GetString("id")

	var reader *bytes.Reader

	// if gin.Mode() == "release" {
	// 	reqBody, err := io.ReadAll(c.Request.Body)
	// 	if err != nil {
	// 		ErrorExit(c, http.StatusInternalServerError, "couldn't read the request body", err)
	// 		return
	// 	}
	//
	// 	var info map[string]any
	// 	err = json.Unmarshal(reqBody, &info)
	// 	if err != nil {
	// 		ErrorExit(c, http.StatusInternalServerError, "couldn't unmarshal the request body", err)
	// 		return
	// 	}
	//
	// 	info["commission_type"] = "bps"
	// 	info["commission"] = "15"
	//
	// 	reqBody, err = json.Marshal(info)
	// 	if err != nil {
	// 		ErrorExit(c, http.StatusInternalServerError, "couldn't marshal the request body again", err)
	// 		return
	// 	}
	//
	// 	reader = bytes.NewReader(reqBody)
	// }

	headers := BasicAuth()

	errs := map[int]string{
		400: "Malformed input",
		403: "Request is forbidden",
		404: "Resource doesn't exist",
		422: "Some parameters are invalid",
	}

	var body map[string]any
	var err error
	if reader != nil {
		body, err = SendRequest[map[string]any](http.MethodPost, BaseURL+Trading+id+"/orders", reader, errs, headers)
		log.Println("Reader")
	} else {
		body, err = SendRequest[map[string]any](http.MethodPost, BaseURL+Trading+id+"/orders", c.Request.Body, errs, headers)
		log.Println("Req body")
	}

	if err != nil {
		RequestExit(c, body, err, "couldn't place an order for the given stock")
		return
	}

	orderID := body["id"].(string)
	createdAt := body["created_at"].(string)
	updatedAt := body["updated_at"].(string)
	symbol := body["symbol"].(string)
	side := body["side"].(string)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err)
		return
	}
	defer conn.Close(context.Background())

	if err = CreateOrdersTable(conn); err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't create a table for the orders", err)
		return
	}

	_, err = conn.Exec(context.Background(), "insert into orders (id, user_id, symbol, side, created_at, updated_at) values ($1, $2, $3, $4, $5, $6)", orderID, id, symbol, side, createdAt, updatedAt)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't put the information about your order in the database", err)
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetOrders(c *gin.Context) {
	id := c.GetString("id")

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err)
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id, symbol, side, created_at, updated_at from orders where user_id = $1", id)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't get the information for the orders from the database", err)
		return
	}

	orders, err := pgx.CollectRows(rows, func(rows pgx.CollectableRow) (*Order, error) {
		o := Order{}
		o.UserID = id
		err := rows.Scan(&o.ID, &o.Symbol, &o.Side, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}

		return &o, nil
	})
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't work with the information from the database", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func GetOrdersAlpaca(c *gin.Context) {
	id := c.GetString("id")
	status := c.Query("status")
	if status == "" {
		status = "open"
	}

	headers := BasicAuth()

	errs := map[int]string{
		400: "Malformed input",
		404: "Resource doesn't exist",
	}

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+"/orders?status="+status, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "couldn't get the orders for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func ReplaceOrder(c *gin.Context) {
	id := c.GetString("id")
	orderID := c.Param("orderId")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Malformed input",
		404: "Resource doesn't exist",
	}

	body, err := SendRequest[any](http.MethodPatch, BaseURL+Trading+id+"/orders/"+orderID, c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "couldn't replce the order")
		return
	}

	c.JSON(http.StatusOK, body)
}

type result struct {
	Type string
	F    func()
}

func CancelOrder(c *gin.Context) {
	id := c.GetString("id")
	orderID := c.Param("orderId")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Malformed input",
		404: "Resource doesn't exist",
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	res := make(chan result)
	var resBody any
	go func() {
		body, err := SendRequest[any](http.MethodDelete, BaseURL+Trading+id+"/orders/"+orderID, c.Request.Body, errs, headers)
		if err != nil {
			res <- result{Type: "f", F: func() {
				RequestExit(c, body, err, "couldn't cancel the order")
			}}
			wg.Done()
			return
		}

		resBody = body
		res <- result{Type: "s"}
		wg.Done()
	}()

	go func() {
		conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			res <- result{Type: "f", F: func() {
				ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err)
			}}
			wg.Done()
			return
		}
		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "delete from orders where id = $1 and user_id = $2", orderID, id)
		if err != nil {
			res <- result{Type: "f", F: func() {
				ErrorExit(c, http.StatusInternalServerError, "couldn't delete the information from the database", err)
			}}
			wg.Done()
			return
		}

		res <- result{Type: "s"}
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(res)
	}()

	for r := range res {
		if r.Type == "f" {
			r.F()
			return
		}
	}

	c.JSON(http.StatusOK, resBody)
}

func EstimateOrder(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodPost, BaseURL+Trading+id+"/orders/estimation", c.Request.Body, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "couldn't estimate the order")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetOrderByID(c *gin.Context) {
	id := c.GetString("id")
	orderID := c.Param("orderId")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Malformed input",
		404: "Resource doesn't exist",
	}

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+"/orders/"+orderID, c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "couldn't get the order")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetAccountProtfolioHistory(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+"/account/portfolio/history", c.Request.Body, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "couldn't get the order")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetOpenPositions(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+"/positions", nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the open positions for your account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func CloseAllOpenPositions(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	errs := map[int]string{
		500: "Failed to liquidate some positions",
	}

	body, err := SendRequest[any](http.MethodDelete, BaseURL+Trading+id+"/positions", nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't close all the open positions for your account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetOpenPosition(c *gin.Context) {
	id := c.GetString("id")
	symbolOrAssetID := c.Param("symbol_or_asset_id")
	if symbolOrAssetID == "" {
		ErrorExit(c, http.StatusBadRequest, "missing a parameter", nil)
		return
	}

	headers := BasicAuth()

	errs := map[int]string{
		404: "Account doesn't have a position for this symbol or asset_id ",
	}

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+"/positions/"+symbolOrAssetID, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the open position for your account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func ClosePosition(c *gin.Context) {
	id := c.GetString("id")
	symbolOrAssetID := c.Param("symbol_or_asset_id")
	if symbolOrAssetID == "" {
		ErrorExit(c, http.StatusBadRequest, "missing a parameter", nil)
		return
	}
	qtyFl := c.GetFloat64("qty")
	qty := int(qtyFl)
	percentageFl := c.GetFloat64("percentage")
	percentage := int(percentageFl)

	if qty != 0 && percentage != 0 {
		ErrorExit(c, http.StatusBadRequest, "only 1 of the bonus parameters can be specified at a time", nil)
		return
	}

	url := BaseURL + Trading + id + "/positions/" + symbolOrAssetID
	if qty != 0 {
		url += "?qty=" + fmt.Sprintf("%d", qty)
	} else if percentage != 0 {
		url += "?percentage=" + fmt.Sprintf("%d", percentage)
	}
	log.Println(url)

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodDelete, url, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the open position for your account")
		return
	}

	c.JSON(http.StatusOK, body)
}
