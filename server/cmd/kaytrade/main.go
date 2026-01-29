package main

import (
	"os"

	"github.com/Phantomvv1/KayTrade/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	if os.Getenv("KAYTRADE_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	if gin.Mode() == "release" {
		//This will stay commented for now since I don't have a domain and a secure http connection
		// Domain = "kay_trade.com" //example domain
		// Secure = true
	}

	r := routes.NewRouter()

	r.Run(":42069")
}
