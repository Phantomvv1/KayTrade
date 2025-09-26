package routes

import (
	"net/http"

	. "github.com/Phantomvv1/KayTrade/internal/auth"
	"github.com/Phantomvv1/KayTrade/internal/clock"
	"github.com/Phantomvv1/KayTrade/internal/documents"
	"github.com/Phantomvv1/KayTrade/internal/journals"
	marketdata "github.com/Phantomvv1/KayTrade/internal/market_data"
	. "github.com/Phantomvv1/KayTrade/internal/middleware"
	"github.com/Phantomvv1/KayTrade/internal/trading"
	"github.com/Phantomvv1/KayTrade/internal/watchlist"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.Any("/", func(c *gin.Context) { c.JSON(http.StatusOK, nil) })
	r.POST("/signup", SignUp)
	r.POST("/login", LogIn)
	r.GET("/profile", AuthParserMiddleware, AuthProtectMiddleware, GetCurrentProfile)
	r.GET("/users", AuthParserMiddleware, AdminOnlyMiddleware, GetAllUsers)
	r.GET("/users/:id", AuthParserMiddleware, AuthProtectMiddleware, GetUser)
	r.GET("/users/alpaca", AuthParserMiddleware, AdminOnlyMiddleware, GetAllUsersAlpaca)
	r.PATCH("/users/:id", AuthParserMiddleware, JSONParserMiddleware, AuthProtectMiddleware, UpdateUser)
	r.PATCH("/users/:id/alpaca", AuthParserMiddleware, AuthProtectMiddleware, UpdateUserAlpaca)
	r.DELETE("/users/:id", AuthParserMiddleware, AuthProtectMiddleware, DeleteUser)
	r.POST("/refresh", JSONParserMiddleware, Refresh)
	r.GET("/clock", clock.GetClock)

	accounts := r.Group("/accounts/:id")
	accounts.Use(AuthParserMiddleware)
	accounts.Use(AuthProtectMiddleware)
	f := accounts.Group("/funding")
	f.POST("", CreateBankRelationship)
	f.GET("", GetBankRelationships)
	f.GET("/alpaca", GetBankRelationshipsAlpaca)
	f.DELETE("", JSONParserMiddleware, DeleteBankRelationship)

	t := accounts.Group("/transfers")
	t.GET("", GetAllTransfers)
	t.POST("", NewTransfer)

	trade := accounts.Group("/trading")
	trade.POST("", trading.CreateOrder)
	trade.GET("", trading.GetOrders)
	trade.GET("/alpaca", trading.GetOrdersAlpaca)
	trade.PATCH("/orders/:orderId", trading.ReplaceOrder)
	trade.DELETE("/orders/:orderId", trading.CancelOrder)
	trade.POST("/orders/estimation", trading.EstimateOrder)
	trade.GET("/orders/:orderId", trading.GetOrderByID)
	trade.GET("/portfolio", trading.GetAccountProtfolioHistory)
	trade.GET("/positions", trading.GetOpenPositions)
	trade.DELETE("/positions", trading.CloseAllOpenPositions)
	trade.GET("/positions/:symbol_or_asset_id", trading.GetOpenPosition)
	trade.DELETE("/positions/:symbol_or_asset_id", JSONParserMiddleware, trading.ClosePosition)

	docs := accounts.Group("/documents")
	docs.GET("", documents.GetAllDocuments)
	docs.GET("/download/:documentId", documents.DownloadDocument)

	journ := accounts.Group("/journals")
	journ.POST("", journals.CreateJournal)
	journ.GET("", journals.GetJournalList)
	journ.DELETE("/:journal_id", journals.CancelJournal)
	journ.GET("/:journal_id", journals.GetJournalByID)

	watch := accounts.Group("/watchlist")
	watch.POST("/alpaca", watchlist.CreateWatchlistAlpaca)
	watch.GET("/alpaca", watchlist.GetWatchlistAlpaca)
	watch.GET("/:watchlistId", watchlist.ManageWatchlistAlpaca)
	watch.PUT("/:watchlistId", watchlist.UpdateWatchlistAlpaca)
	watch.DELETE("/:watchlistId", watchlist.DeleteWatchlistAlpaca)
	watch.POST("alpaca/:watchlistId", watchlist.AddAssetWatchlistAlpaca)
	watch.DELETE("/:watchlistId/:symbol", watchlist.RemoveSymbolFromWatchlistAlpaca)
	watch.POST("/:symbol", watchlist.AddSymbolToWatchlist)
	watch.GET("", watchlist.GetSymbolsFromWatchlist)

	data := r.Group("/data")
	data.GET("/auctions", SymbolsParserMiddleware, marketdata.GetHistoricalAuctions)
	data.GET("/bars", SymbolsParserMiddleware, StartParserMiddleware, marketdata.GetHistoricalBars)
	data.GET("/bars/latest", SymbolsParserMiddleware, marketdata.GetLatestBars)
	data.GET("/conditions/:ticktype", marketdata.GetConditionCodes)
	data.GET("/exchanges", marketdata.GetExchangeCodes)
	data.GET("/quotes", SymbolsParserMiddleware, StartParserMiddleware, marketdata.GetHisoticalQuotes)
	data.GET("/quotes/latest", SymbolsParserMiddleware, marketdata.GetLatestQuotes)
	data.GET("/snapshots", SymbolsParserMiddleware, marketdata.GetSnapshots)
	data.GET("/trades", SymbolsParserMiddleware, StartParserMiddleware, marketdata.GetHistoricalTrades)
	data.GET("/trades/latest", SymbolsParserMiddleware, marketdata.GetLatestTrades)
	data.GET("/stocks/most-active", marketdata.GetMostActiveStocks)
	data.GET("/stocks/top-market-movers", marketdata.GetTopMarketMovers)

	hub := marketdata.NewHub()
	go hub.Run()
	data.GET("/stocks/live/:symbol", func(c *gin.Context) {
		marketdata.GetRealTimeStocks(c, hub)
	})

	return r
}
