package auth

import (
	"context"
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

type Bank struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Type      string     `json:"type"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Timer `json:"updated_at"`
}

func CreateBankTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists bank(id uuid primary key, user_id uuid references authentication(id) on delete cascade, "+
		"type text, created_at timestamp default current_timestamp, updated_at timestamp default current_timestamp)")
	return err
}

func CreateBankRelationship(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Bad request",
		409: "A bank relationship already exists for this account",
	}

	body, err := SendRequest[map[string]string](http.MethodPost, BaseURL+Accounts+id+"/recipient_banks", c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to create a bank relationship")
		return
	}

	bankID := body["id"]
	log.Println(bankID)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err)
		return
	}
	defer conn.Close(context.Background())

	if err = CreateBankTable(conn); err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't create a table for bank relationships", err)
		return
	}

	_, err = conn.Exec(context.Background(), "insert into bank (id, user_id, type) values ($1, $2, 'bank')", bankID, id)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't put your information into the database", err)
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetBankRelationships(c *gin.Context) {
	id := c.GetString("id")

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err)
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id, user_id, type, created_at, updated_at from bank b where b.user_id = $1", id)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't get the information from the database", err)
		return
	}

	var banks []Bank
	for rows.Next() {
		bank := Bank{}
		err = rows.Scan(&bank.ID, &bank.UserID, &bank.Type, &bank.CreatedAt, &bank.UpdatedAt)
		if err != nil {
			ErrorExit(c, http.StatusInternalServerError, "couldn't work with the items from the database", err)
			return
		}

		banks = append(banks, bank)
	}

	if rows.Err() != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't get the items from the database", rows.Err())
		return
	}

	c.JSON(http.StatusOK, gin.H{"relationships": banks})
}

// This is an alpaca endpoint. Only use it if you want more information about the
// relationships or the banks.
func GetBankRelationshipsAlpaca(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Bad request. The body in the request is not valid.",
	}

	body, err := SendRequest[any](http.MethodGet, BaseURL+Accounts+id+"/recipient_banks", nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to get bank relationships for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func DeleteBankRelationship(c *gin.Context) {
	id := c.GetString("id")
	bankID := c.GetString("bank_id")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Bad request",
		404: "No Bank Relationship with the id specified by bank_id was found for this account",
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	type result struct {
		Type string
		F    func()
	}
	res := make(chan result)
	var resBody any
	go func() {
		body, err := SendRequest[any](http.MethodDelete, BaseURL+Accounts+id+"/recipient_banks/"+bankID, c.Request.Body, errs, headers)
		if err != nil {
			res <- result{Type: "r", F: func() { RequestExit(c, body, err, "unable to delete the bank relationships for this account") }}
			wg.Done()
			return
		}

		resBody = body

		res <- result{Type: "s", F: func() {}}
		wg.Done()
	}()

	go func() {
		conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			res <- result{Type: "n", F: func() { ErrorExit(c, http.StatusInternalServerError, "couldn't connect to the database", err) }}
			wg.Done()
			return
		}
		defer conn.Close(context.Background())

		_, err = conn.Exec(context.Background(), "delete from bank where user_id = $1 and id = $2", bankID, id)
		if err != nil {
			res <- result{Type: "n", F: func() {
				ErrorExit(c, http.StatusInternalServerError, "coludn't delete the information from the database", err)
			}}
			wg.Done()
			return
		}

		res <- result{Type: "s", F: func() {}}
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(res)
	}()

	for r := range res {
		if r.Type != "s" {
			r.F()
			return
		}
	}

	c.JSON(http.StatusOK, resBody)
}

func CreateAchRelationship(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Malformed input",
		401: "Client is not authorized for this operation",
		409: "The account already has an active ach relationship",
	}

	body, err := SendRequest[map[string]any](http.MethodPost, BaseURL+Accounts+id+"/ach_relationships", c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to create an ach relationship for this account")
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "insert into bank (id, user_id, type) values ($1, $2, 'ach')", body["id"].(string), id)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "coludn't delete the information from the database", err)
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetAchRelationships(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Accounts+id+"/ach_relationships", c.Request.Body, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to get the ach relationship for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func DeleteAchRelationship(c *gin.Context) {
	id := c.GetString("id")
	relationshipID := c.GetString("relationshipID")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Malformed input",
		401: "Client is not authorized for this operation",
		409: "The account already has an active ach relationship",
	}

	body, err := SendRequest[map[string]any](http.MethodDelete, BaseURL+Accounts+id+"/ach_relationships/"+relationshipID, c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to create an ach relationship for this account")
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "delete from bank where user_id = $1 and id = $2", id, relationshipID)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "coludn't delete the information from the database", err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func GetAllTransfers(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Accounts+id+"/transfers", nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to get the transfers for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func NewTransfer(c *gin.Context) {
	id := c.GetString("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodPost, BaseURL+Accounts+id+"/transfers", c.Request.Body, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to get the transfers for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}
