package main

import (
	"log"
	"net/http"

	. "github.com/Phantomvv1/KayTrade/internal/auth"
	. "github.com/Phantomvv1/KayTrade/internal/middleware"
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
	r.GET("/profile", AuthMiddleware, GetCurrentProfile)
	r.GET("/users", AuthMiddleware, GetAllUsers)
	r.POST("/refresh", AuthMiddleware, Refresh)

	r.Run(":42069")
}
