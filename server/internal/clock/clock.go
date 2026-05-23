package clock

import (
	"net/http"
	"strings"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
)

func GetClock(c *gin.Context) {
	headers := BasicAuth()
	marketsArr := c.QueryArray("markets")
	markets := strings.Join(marketsArr, ",")

	baseUrl := []byte(BaseURL)
	baseUrl[len(baseUrl)-2] = '2'
	body, err := SendRequest[any](http.MethodGet, string(baseUrl)+Clock+"?markets="+markets, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the clock")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetCalendar(c *gin.Context) {
	headers := BasicAuth()
	market := c.Param("market")
	timezone := c.Query("timezone")

	baseUrl := []byte(BaseURL)
	baseUrl[len(baseUrl)-2] = '2'

	var body any
	var err error
	if timezone != "" {
		body, err = SendRequest[any](http.MethodGet, string(baseUrl)+Calendar+market+"?timezone="+timezone, nil, nil, headers)
	} else {
		body, err = SendRequest[any](http.MethodGet, string(baseUrl)+Calendar+market, nil, nil, headers)
	}
	if err != nil {
		RequestExit(c, body, err, "coludn't get the clock")
		return
	}

	c.JSON(http.StatusOK, body)
}
