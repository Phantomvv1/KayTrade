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
	r.POST("/sign-up", SignUp)
	r.POST("/log-in", LogIn)
	r.POST("/refresh", JSONParserMiddleware, Refresh)
	r.GET("/clock", clock.GetClock)
	r.GET("/search", AuthMiddleware, watchlist.SearchCompanies)
	r.GET("/company-information/:symbol", AuthMiddleware, watchlist.GetCompanyInformation)

	users := r.Group("/users")
	users.Use(AuthMiddleware)
	users.GET("", GetUser)
	users.GET("/alpaca", GetUserAlpaca)
	users.GET("/all", AdminOnlyMiddleware, GetAllUsers)
	users.GET("/all/alpaca", AdminOnlyMiddleware, GetAllUsersAlpaca)
	users.GET("/trading-details", GetAccountTradingDetails)
	users.PATCH("", JSONParserMiddleware, UpdateUser)
	users.PATCH("/alpaca", UpdateUserAlpaca)
	users.DELETE("", DeleteUser)

	f := r.Group("/funding")
	f.Use(AuthMiddleware)
	f.POST("", CreateBankRelationship)
	f.POST("/ach", CreateAchRelationship)
	f.GET("/ach", GetAchRelationship)
	f.GET("", GetBankRelationships)
	f.GET("/alpaca", GetBankRelationshipsAlpaca)
	f.DELETE("", JSONParserMiddleware, DeleteBankRelationship)
	f.DELETE("ach", DeleteAchRelationship)

	t := r.Group("/transfers")
	t.Use(AuthMiddleware)
	t.GET("", GetAllTransfers)
	t.POST("", NewTransfer)

	trade := r.Group("/trading")
	trade.Use(AuthMiddleware)
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

	docs := r.Group("/documents")
	docs.Use(AuthMiddleware)
	docs.GET("", documents.GetAllDocuments)
	docs.GET("/download/:documentId", documents.DownloadDocument)

	journ := r.Group("/journals")
	journ.Use(AuthMiddleware)
	journ.POST("", journals.CreateJournal)
	journ.GET("", journals.GetJournalList)
	journ.DELETE("/:journal_id", journals.CancelJournal)
	journ.GET("/:journal_id", journals.GetJournalByID)

	watch := r.Group("/watchlist")
	watch.Use(AuthMiddleware)
	watch.POST("/alpaca", watchlist.CreateWatchlistAlpaca)
	watch.GET("/alpaca", watchlist.GetWatchlistAlpaca)
	watch.GET("/alpaca/:watchlistId", watchlist.ManageWatchlistAlpaca)
	watch.PUT("/alpaca/:watchlistId", watchlist.UpdateWatchlistAlpaca)
	watch.DELETE("/alpaca/:watchlistId", watchlist.DeleteWatchlistAlpaca)
	watch.POST("/alpaca/:watchlistId", watchlist.AddAssetWatchlistAlpaca)
	watch.DELETE("/alpaca/:watchlistId/:symbol", watchlist.RemoveSymbolFromWatchlistAlpaca)
	watch.POST("/:symbol", watchlist.AddSymbolToWatchlist)
	watch.GET("", watchlist.GetSymbolsFromWatchlist)
	watch.GET("/info", watchlist.GetInformationForSymbols)
	watch.DELETE("/:symbol", watchlist.RemoveSymbolFromWatchlist)
	watch.DELETE("", watchlist.RemoveAllSymbolsFromWatchlist)

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
