package marketdata

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
)

type TimeFrame string

func (t TimeFrame) ValidTimeFrame() bool {
	index := 0
	for i, char := range t {
		if char < '0' || char > '9' {
			index = i
			break
		}
	}

	number := string([]byte(t)[:index])
	switch []byte(t)[index] {
	case 'T':
		n, err := strconv.Atoi(number)
		if err != nil {
			return false
		}

		if n < 0 || n > 59 {
			return false
		}

		return true
	case 'H':
		n, err := strconv.Atoi(number)
		if err != nil {
			return false
		}

		if n < 0 || n > 23 {
			return false
		}
		return true
	case 'D':
		return true
	case 'W':
		return true
	case 'M':
		switch number {
		case "1":
			return true
		case "2":
			return true
		case "3":
			return true
		case "4":
			return true
		case "6":
			return true
		case "12":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

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

	// Getting the last market data if today's market hasn't opened
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

func GetHistoricalBars(c *gin.Context) {
	symbols := c.QueryArray("symbols")
	if symbols == nil {
		ErrorExit(c, http.StatusBadRequest, "no information given", nil)
		return
	} else if symbols[0] == "" {
		ErrorExit(c, http.StatusBadRequest, "no information given", nil)
		return
	}

	symbolsToSend := strings.Join(symbols, ",")

	timeframe := TimeFrame(c.Query("timeframe"))
	if timeframe == "" || !timeframe.ValidTimeFrame() {
		ErrorExit(c, http.StatusBadRequest, "timeframe was incorrectly provided", nil)
		return
	}

	start := c.Query("start")
	omitInterval := false
	if start == "" {
		omitInterval = true
	} else {
		start = "&start=" + start
	}

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	var body any
	var err error
	if omitInterval {
		body, err = SendRequest[any](http.MethodGet, MarketData+"/stocks/bars?symbols="+symbolsToSend+"&timeframe="+string(timeframe), nil, errs, headers)
	} else {
		body, err = SendRequest[any](http.MethodGet, MarketData+"/stocks/bars?symbols="+symbolsToSend+"&timeframe="+string(timeframe)+start, nil, errs, headers)
	}
	if err != nil {
		RequestExit(c, body, err, "coludn't get the market data for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetLatestBars(c *gin.Context) {
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

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/bars/latest?symbols="+symbolsToSend, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the market data for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetConditionCodes(c *gin.Context) {
	ticktype := c.Param("ticktype")
	if ticktype == "" {
		ErrorExit(c, http.StatusBadRequest, "incorrectly provided tick type", nil)
		return
	}

	if ticktype != "trade" && ticktype != "quote" {
		ErrorExit(c, http.StatusBadRequest, "incorrectly provided tick type", nil)
		return
	}

	tape := c.Query("tape")
	if tape != "A" && tape != "B" && tape != "C" {
		ErrorExit(c, http.StatusBadRequest, "incorrectly provided tape", nil)
		return
	}

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/meta/coditions/"+ticktype+"?tape="+tape, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the market data for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetExchangeCodes(c *gin.Context) {
	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/meta/exchanges", nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the market data for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetHisoticalQuotes(c *gin.Context) {
	symbols := c.QueryArray("symbols")
	if symbols == nil {
		ErrorExit(c, http.StatusBadRequest, "no information given", nil)
		return
	} else if symbols[0] == "" {
		ErrorExit(c, http.StatusBadRequest, "no information given", nil)
		return
	}

	symbolsToSend := strings.Join(symbols, ",")

	start := c.Query("start")
	omitInterval := false
	if start == "" {
		omitInterval = true
	} else {
		start = "&start=" + start
	}

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	var body any
	var err error
	if omitInterval {
		body, err = SendRequest[any](http.MethodGet, MarketData+"/stocks/quotes?symbols="+symbolsToSend, nil, errs, headers)
	} else {
		body, err = SendRequest[any](http.MethodGet, MarketData+"/stocks/quotes?symbols="+symbolsToSend+start, nil, errs, headers)
	}
	if err != nil {
		RequestExit(c, body, err, "coludn't get the market data for these symbols")
		return
	}
	if err != nil {
		RequestExit(c, body, err, "coludn't get the market data for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}
