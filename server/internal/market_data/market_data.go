package marketdata

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader websocket.Upgrader

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

type User struct {
	Symbol string
	ws     *websocket.Conn
	send   chan map[string]any
}

func (u User) Read(hub *Hub) {
	for {
		_, message, err := u.ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if string(message) == "exit" {
			hub.Unregister <- &u
		}
	}
}

func (u User) Write(hub *Hub) {
	for data := range <-u.send {
		err := u.ws.WriteJSON(data)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

type Message struct {
	Receiver string
	Message  string
	Symbol   string
	Data     map[string]any
}

type Hub struct {
	Users       map[*User]struct{}
	Broadcast   chan *Message
	Register    chan *User
	Unregister  chan *User
	IsConnected bool
	IsListening bool
	ws          *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		Users:       make(map[*User]struct{}),
		Broadcast:   make(chan *Message),
		Register:    make(chan *User),
		Unregister:  make(chan *User),
		IsConnected: false,
		IsListening: false,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case msg := <-h.Broadcast:
			switch msg.Receiver {
			case "all":
				for user := range h.Users {
					user.send <- msg.Data
				}
			case "":
				for user := range h.Users {
					if user.Symbol == msg.Symbol {
						user.send <- msg.Data
					}
				}
			default:
			}

		case user := <-h.Register:
			h.Users[user] = struct{}{}
			if !h.IsConnected {
				h.IsConnected = true
				go h.Connect(user.Symbol)
			}

			alreadySubscribed := false
			for existingUser := range h.Users {
				if user.Symbol == existingUser.Symbol {
					alreadySubscribed = true
				}
			}

			if !alreadySubscribed {
				go h.Subscribe(user.Symbol)
			}

		case user := <-h.Unregister:
			if _, ok := h.Users[user]; ok {
				delete(h.Users, user)
				close(user.send)
			}
		}
	}
}

func (h *Hub) Connect(symbol string) {
	ws, _, err := websocket.DefaultDialer.Dial(RealTimeData, nil)
	if err != nil {
		h.Broadcast <- &Message{Receiver: "all", Message: "Error dialing the data stream"}
		return
	}
	h.ws = ws

	var body []map[string]string
	if err = ws.ReadJSON(&body); err != nil {
		h.Broadcast <- &Message{Receiver: "all", Message: "Error dialing the data stream"}
		return
	}

	if body[0]["T"] != "success" && body[0]["msg"] != "connected" {
		h.Broadcast <- &Message{Receiver: "all", Message: "Error couldn't connect to the real time data stream. Please check if the market is open. " +
			"If it's not, please wait for it. Otherwise try again."}
		return
	}

	authMsg := map[string]string{
		"action": "auth",
		"key":    os.Getenv("API_KEY"),
		"secret": os.Getenv("SECRET_KEY"),
	}

	if err = ws.WriteJSON(authMsg); err != nil {
		h.Broadcast <- &Message{Receiver: "all", Message: "Error writing in the data stream"}
		return
	}

	if err = ws.ReadJSON(&body); err != nil {
		h.Broadcast <- &Message{Receiver: "all", Message: "Error reading in the data stream"}
		return
	}

	if body[0]["T"] != "success" && body[0]["msg"] != "authenticated" {
		h.Broadcast <- &Message{Receiver: "all", Message: "Error couldn't connect to the real time data stream. Please check if the market is open. " +
			"If it's not, please wait for it. Otherwise try again."}
		return
	}

	h.Subscribe(symbol)
}

func (h *Hub) Subscribe(symbol string) {
	if symbol == "" {
		return
	}

	body := map[string]string{
		"action":   "subscribe",
		"trades":   symbol,
		"quotes":   symbol,
		"bars":     symbol,
		"statuses": "*",
	}

	if err := h.ws.WriteJSON(body); err != nil {
		h.Broadcast <- &Message{Receiver: "all", Message: "Error couldn't subscribe to these symbols"}
		return
	}

	var resp []map[string]any
	if err := h.ws.ReadJSON(&resp); err != nil {
		h.Broadcast <- &Message{Receiver: "all", Message: "Error couldn't subscribe to these symbols"}
		return
	}

	if resp[0]["T"] != "subscription" {
		h.Broadcast <- &Message{Receiver: "all", Message: "Error couldn't subscribe to these symbols"}
		return
	}

	if !h.IsListening {
		h.IsListening = true
		go h.Listen()
	}
}

func (h *Hub) Listen() {
	for {
		var body []map[string]any
		if err := h.ws.ReadJSON(&body); err != nil {
			h.Broadcast <- &Message{Receiver: "all", Message: "Error couldn't subscribe to these symbols"}
			return
		}

		h.Broadcast <- &Message{Receiver: "", Message: "", Symbol: body[0]["S"].(string), Data: body[0]}
	}
}

func GetHistoricalAuctions(c *gin.Context) {
	symbols := c.GetString("symbols")

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
		body, err = SendRequest[any](http.MethodGet, MarketData+"/stocks/auctions?symbols="+symbols, nil, errs, headers)
	} else {
		body, err = SendRequest[any](http.MethodGet, MarketData+"/stocks/auctions?symbols="+symbols+start+end, nil, errs, headers)
	}
	if err != nil {
		RequestExit(c, body, err, "coludn't get the market data for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetHistoricalBars(c *gin.Context) {
	symbols := c.GetString("symbols")
	start := c.GetString("start")

	timeframe := TimeFrame(c.Query("timeframe"))
	if timeframe == "" || !timeframe.ValidTimeFrame() {
		ErrorExit(c, http.StatusBadRequest, "timeframe was incorrectly provided", nil)
		return
	}

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/bars?symbols="+symbols+start+"&timeframe="+string(timeframe), nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the market data for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetLatestBars(c *gin.Context) {
	symbols := c.GetString("symbols")

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/bars/latest?symbols="+symbols, nil, errs, headers)
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
	symbols := c.GetString("symbols")
	start := c.GetString("start")

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/quotes?symbols="+symbols+start, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the qoutes for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetLatestQuotes(c *gin.Context) {
	symbols := c.GetString("symbols")

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/quotes/latest?symbols="+symbols, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the qoutes for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetSnapshots(c *gin.Context) {
	symbols := c.GetString("symbols")

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/snapshots?symbols="+symbols, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the qoutes for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetHistoricalTrades(c *gin.Context) {
	symbols := c.GetString("symbols")
	start := c.GetString("start")

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/trades?symbols="+symbols+start, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the qoutes for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetLatestTrades(c *gin.Context) {
	symbols := c.GetString("symbols")

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, MarketData+"/stocks/trades/latest?symbols="+symbols, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the qoutes for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetMostActiveStocks(c *gin.Context) {
	by := c.Query("by")
	top := c.Query("top")

	if by == "" {
		by = "volume"
	} else if by != "volume" && by != "trades" {
		ErrorExit(c, http.StatusBadRequest, "incorrectly provided by", nil)
		return
	}

	if top == "" {
		top = "10"
	} else if top < "0" || top > "100" {
		ErrorExit(c, http.StatusBadRequest, "incorrectly provided top", nil)
		return
	}

	by = "?by=" + by
	top = "&top=" + top

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, "https://data.sandbox.alpaca.markets/v1beta1/screener/stocks/most-actives"+by+top, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the qoutes for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetTopMarketMovers(c *gin.Context) {
	top := c.Query("top")

	if top == "" {
		top = "10"
	} else if top < "0" || top > "50" {
		ErrorExit(c, http.StatusBadRequest, "incorrectly provided top", nil)
		return
	}

	headers := BasicAuth()

	errs := map[int]string{
		400: "One of the request parameters is invalid",
		403: "Authentication headers are missing or invalid. Make sure you authenticate your request with a valid API key",
		429: "Too many requests",
		500: "Internal server error. We recommend retrying these later",
	}

	body, err := SendRequest[any](http.MethodGet, "https://data.sandbox.alpaca.markets/v1beta1/screener/stocks/movers?top="+top, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the qoutes for these symbols")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetRealTimeStocks(c *gin.Context, hub *Hub) {
	symbol := c.Param("symbol")

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		ErrorExit(c, http.StatusInternalServerError, "couldn't upgrade the connection to a websocket one", err)
		return
	}

	user := &User{Symbol: symbol, ws: ws, send: make(chan map[string]any)}
	hub.Register <- user

	go user.Read(hub)
	go user.Write(hub)
}
