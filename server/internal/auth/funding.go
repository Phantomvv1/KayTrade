package auth

import (
	"net/http"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
)

func CreateBankRelationship(c *gin.Context) {
	id := c.Param("id")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Bad request",
		409: "A bank relationship already exists for this account",
	}

	body, err := SendRequest[any](http.MethodPost, BaseURL+Accounts+id+"/recipient_banks", c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to create a bank relationship")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetBankRelationships(c *gin.Context) {
	id := c.Param("id")

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
	id := c.Param("id")
	bankID := c.GetString("bank_id")

	headers := BasicAuth()

	errs := map[int]string{
		400: "Bad request",
		404: "No Bank Relationship with the id specified by bank_id was found for this account",
	}

	body, err := SendRequest[any](http.MethodDelete, BaseURL+Accounts+id+"/recipient_banks/"+bankID, c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to delete the bank relationships for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetAllTransfers(c *gin.Context) {
	id := c.Param("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Accounts+id+"/transfers", nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "unable to get the transfers for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func NewTransfer(c *gin.Context) {

}
