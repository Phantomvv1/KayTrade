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
	r.GET("/profile", AuthParserMiddleware, GetCurrentProfile)
	r.GET("/users", AuthParserMiddleware, AuthProtectMiddleware, GetAllUsers)
	r.GET("/users/:id", AuthParserMiddleware, GetUser)
	r.GET("/users/alpaca", AuthParserMiddleware, AuthProtectMiddleware, GetAllUsersAlpaca)
	r.PATCH("/users/:id", AuthParserMiddleware, JSONParserMiddleware, UpdateUser)
	r.PATCH("/users/:id/alpaca", AuthParserMiddleware, UpdateUserAlpaca)
	r.POST("/refresh", JSONParserMiddleware, Refresh)

	r.Run(":42069")
}
