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

	r.Run(":42069")
}
