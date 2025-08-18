package marketdata

import (
	"net/http"
	"strings"
	"time"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
)

func GetHistoricalAuctions(c *gin.Context) {
	symbols := c.QueryArray("symbols")
	if symbols == nil {
		ErrorExit(c, http.StatusBadRequest, "no information given", nil)
		return
	} else if symbols[0] == "" {
		ErrorExit(c, http.StatusBadRequest, "no information given", nil)
		return
	}

	symbolsToSend := strings.Join(symbols, ",")

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	omitDuration := true
	now := time.Now().UTC()
	start := "&start="
	end := "&end="
	if now.Hour() < 13 || now.Hour() >= 20 { // market opens at 13:30 UTC and closes at 20:00 UTC
		if now.Weekday() == time.Monday {
			start += now.AddDate(0, 0, -3).Truncate(time.Hour * 24).Format(time.RFC3339)
			end += now.AddDate(0, 0, -2).Truncate(time.Hour * 24).Format(time.RFC3339)
		} else {
			start += now.AddDate(0, 0, -1).Truncate(time.Hour * 24).Format(time.RFC3339)
			end += now.Truncate(time.Hour * 24).Format(time.RFC3339)
		}
	}
	if now.Hour() == 13 && now.Minute() < 30 {
		if now.Weekday() == time.Monday {
			start += now.AddDate(0, 0, -3).Truncate(time.Hour * 24).Format(time.RFC3339)
			end += now.AddDate(0, 0, -2).Truncate(time.Hour * 24).Format(time.RFC3339)
		} else {
			start += now.AddDate(0, 0, -1).Truncate(time.Hour * 24).Format(time.RFC3339)
			end += now.Truncate(time.Hour * 24).Format(time.RFC3339)
		}
	} else {
		omitDuration = false
	}

	var body any
	var err error
	if omitDuration {
		body, err = SendRequest[any](http.MethodGet, MarketData+"/stocks/auctions?symbols="+symbolsToSend, nil, errs, headers)
	} else {
		body, err = SendRequest[any](http.MethodGet, MarketData+"/stocks/auctions?symbols="+symbolsToSend+start+end, nil, errs, headers)
	}
	if err != nil {
		RequestExit(c, body, err, "coludn't get the market data for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}
