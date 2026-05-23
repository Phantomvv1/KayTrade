package clock

import (
	"errors"
	"net/http"
	"strings"
	"time"

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
	start := c.Query("start")
	end := c.Query("end")

	baseUrl := []byte(BaseURL)
	baseUrl[len(baseUrl)-2] = '2'

	writtenQuestionMark := false
	urlToReach := string(baseUrl) + Calendar + market
	if timezone != "" {
		writtenQuestionMark = true
		urlToReach += "?timezone=" + timezone
	}

	if start != "" {
		if !writtenQuestionMark {
			urlToReach += "?start=" + start
		}

		urlToReach += "&start=" + start
	}

	if end != "" {
		if !writtenQuestionMark {
			urlToReach += "?end=" + end
		}

		urlToReach += "&end=" + end
	}

	body, err := SendRequest[any](http.MethodGet, urlToReach, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the clock")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetLastMarketOpenDay(market string) (*time.Time, error) {
	headers := BasicAuth()

	baseUrl := []byte(BaseURL)
	baseUrl[len(baseUrl)-2] = '2'

	start := time.Now().UTC().AddDate(0, 0, -14).Format(time.DateOnly)
	end := time.Now().UTC().Format(time.DateOnly)
	body, err := SendRequest[map[string]any](http.MethodGet, string(baseUrl)+Calendar+market+"?timezone=UTC&start="+start+"&end="+end, nil, nil, headers)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	calendarInfo := body["calendar"].([]any)
	for i := range calendarInfo {
		info := calendarInfo[len(calendarInfo)-(i+1)].(map[string]any)
		coreStart := info["core_start"].(string)

		startTs, err := time.Parse(time.RFC3339, coreStart)
		if err != nil {
			return nil, err
		}

		if now.After(startTs) {
			return &startTs, nil
		}
	}

	return nil, errors.New("Error: wasn't able to find the last day the given stock market was open")
}

func IsStockMarketOpen(market string) (bool, error) {
	headers := BasicAuth()

	baseUrl := []byte(BaseURL)
	baseUrl[len(baseUrl)-2] = '2'
	body, err := SendRequest[map[string][]map[string]any](http.MethodGet, string(baseUrl)+Clock+"?markets="+market, nil, nil, headers)
	if err != nil {
		return false, err
	}

	isMarketDay := body["clocks"][0]["is_market_day"].(bool)
	marketPhase := body["clocks"][0]["phase"].(string)

	if !isMarketDay {
		return false, nil
	}

	if marketPhase != "core" {
		return false, nil
	}

	return true, nil
}
