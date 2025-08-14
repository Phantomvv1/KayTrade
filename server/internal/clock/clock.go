package clock

import (
	"net/http"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
)

func GetClock(c *gin.Context) {
	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Clock, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the clock")
		return
	}

	c.JSON(http.StatusOK, body)
}
