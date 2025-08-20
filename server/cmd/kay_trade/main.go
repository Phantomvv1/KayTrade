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
	data.GET("/auctions", marketdata.GetHistoricalAuctions)
	data.GET("/bars", marketdata.GetHistoricalBars)
	data.GET("/bars/latest", marketdata.GetLatestBars)
	data.GET("/conditions/:ticktype", marketdata.GetConditionCodes)
	data.GET("/exchanges", marketdata.GetExchangeCodes)
	data.GET("/quotes", marketdata.GetHisoticalQuotes)

	r.Run(":42069")
}
