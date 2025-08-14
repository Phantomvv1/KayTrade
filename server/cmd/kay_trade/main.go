package main

import (
	"log"
	"net/http"

	. "github.com/Phantomvv1/KayTrade/internal/auth"
	"github.com/Phantomvv1/KayTrade/internal/clock"
	"github.com/Phantomvv1/KayTrade/internal/documents"
	"github.com/Phantomvv1/KayTrade/internal/journals"
	. "github.com/Phantomvv1/KayTrade/internal/middleware"
	"github.com/Phantomvv1/KayTrade/internal/trading"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	if gin.Mode() == "release" {
		log.Println("In release mode")
		Domain = "kay_trade.com" //example domain
		Secure = true
	}

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

	f := r.Group("/funding")
	f.Use(AuthParserMiddleware)
	f.Use(AuthProtectMiddleware)
	f.POST("/:id", CreateBankRelationship)
	f.GET("/:id", GetBankRelationships)
	f.GET("/:id/alpaca", GetBankRelationshipsAlpaca)
	f.DELETE("/:id", JSONParserMiddleware, DeleteBankRelationship)

	t := r.Group("/transfers")
	t.Use(AuthParserMiddleware)
	t.Use(AuthProtectMiddleware)
	t.GET("/:id", GetAllTransfers)
	t.POST("/:id", NewTransfer)

	trade := r.Group("/trading")
	trade.Use(AuthParserMiddleware)
	trade.Use(AuthProtectMiddleware)
	trade.POST("/:id", trading.CreateOrder)
	trade.GET("/:id", trading.GetOrders)
	trade.GET("/:id/alpaca", trading.GetOrdersAlpaca)
	trade.PATCH("/:id/orders/:orderId", trading.ReplaceOrder)
	trade.DELETE("/:id/orders/:orderId", trading.CancelOrder)
	trade.POST("/:id/orders/estimation", trading.EstimateOrder)
	trade.GET("/:id/orders/:orderId", trading.GetOrderByID)
	trade.GET("/:id/portfolio", trading.GetAccountProtfolioHistory)
	trade.GET("/:id/positions", trading.GetOpenPositions)
	trade.DELETE("/:id/positions", trading.CloseAllOpenPositions)
	trade.GET("/:id/positions/:symbol_or_asset_id", trading.GetOpenPosition)
	trade.DELETE("/:id/positions/:symbol_or_asset_id", JSONParserMiddleware, trading.ClosePosition)

	docs := r.Group("/documents")
	docs.Use(AuthParserMiddleware)
	docs.Use(AuthProtectMiddleware)
	docs.GET("/:id", documents.GetAllDocuments)
	docs.GET("/:id/download/:documentId", documents.DownloadDocument)

	journ := r.Group("/journals")
	journ.Use(AuthParserMiddleware)
	journ.Use(AuthProtectMiddleware)
	journ.POST("", journals.CreateJournal)
	journ.GET("", journals.GetJournalList)
	journ.DELETE("/:id", journals.CancelJournal)
	journ.GET("/:id", journals.GetJournalByID)

	r.Run(":42069")
}
