package journals

import (
	"net/http"
	"strings"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
)

func CreateJournal(c *gin.Context) {
	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the parameters is invalid",
		403: "The ammount requested is not available",
		404: "One of the accounts is not found",
	}

	body, err := SendRequest[any](http.MethodPost, BaseURL+Journals, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't make the journal transaction")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetJournalList(c *gin.Context) {
	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the parameters is invalid",
		422: "The result exceeds 100_000 records",
	}

	body, err := SendRequest[any](http.MethodGet, BaseURL+Journals, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the journals")
		return
	}

	c.JSON(http.StatusOK, body)
}

func CancelJournal(c *gin.Context) {
	id := c.Param("journal_id")

	headers := BasicAuth()

	errs := map[int]string{
		404: "The journal is not found",
		422: "The journal is not in pedning status",
	}

	body, err := SendRequest[any](http.MethodDelete, BaseURL+Journals+id, nil, errs, headers)
	if err != nil {
		if err.Error() == errs[404] || err.Error() == errs[422] {
			RequestExit(c, body, err, strings.ToLower(err.Error()))
			return
		}

		RequestExit(c, body, err, "coludn't cancel the journals")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetJournalByID(c *gin.Context) {
	id := c.Param("journal_id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Journals+id, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the journals")
		return
	}

	c.JSON(http.StatusOK, body)
}
