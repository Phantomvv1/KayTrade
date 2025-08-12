package trading

import (
	"context"
	"net/http"
	"os"
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
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateOrdersTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists orders(id uuid primary key, user_id uuid references authentication(id), "+
		"symbol text, created_at timestamp, updated_at timestamp)")
	return err
}

func CreateOrder(c *gin.Context) {
	id := c.Param("id")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Malformed input",
		403: "Request is forbidden",
		404: "Resource doesn't exist",
		422: "Some parameters are invalid",
	}

	body, err := SendRequest[map[string]any](http.MethodPost, BaseURL+Trading+id+"/orders", c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "couldn't place an order for the given stock")
		return
	}

	orderID := body["id"].(string)
	createdAt := body["created_at"].(string)
	updatedAt := body["updated_at"].(string)
	symbol := body["symbol"].(string)

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

	_, err = conn.Exec(context.Background(), "insert into orders (id, user_id, symbol, created_at, updated_at) values ($1, $2, $3, $4, $5)", orderID, id, symbol, createdAt, updatedAt)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't put the information about your order in the database", err)
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetOrders(c *gin.Context) {
	id := c.Param("id")

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err)
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id, symbol, created_at, updated_at from orders where user_id = $1", id)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't get the information for the orders from the database", err)
		return
	}

	orders, err := pgx.CollectRows(rows, func(rows pgx.CollectableRow) (*Order, error) {
		o := Order{}
		o.UserID = id
		err := rows.Scan(&o.ID, &o.Symbol, &o.CreatedAt, &o.UpdatedAt)
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
	id := c.Param("id")
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
