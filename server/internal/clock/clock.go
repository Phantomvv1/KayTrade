package clock

import (
	"net/http"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
)

func GetClock(c *gin.Context) {
	headers := BasicAuth()

	errs := map[int]string{
		404: "Not found",
	}

	body, err := SendRequest[any](http.MethodGet, BaseURL+Clock, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the documents for your account")
		return
	}

	c.JSON(http.StatusOK, body)
}
