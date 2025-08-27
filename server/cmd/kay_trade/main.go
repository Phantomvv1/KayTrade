package main

import (
	. "github.com/Phantomvv1/KayTrade/internal/auth"
	"github.com/Phantomvv1/KayTrade/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	if gin.Mode() == "release" {
		Domain = "kay_trade.com" //example domain
		Secure = true
	}

	r := routes.NewRouter()

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

	r.Run(":42069")
}
