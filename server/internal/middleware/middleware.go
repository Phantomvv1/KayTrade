package middleware

import (
	"log"
	"net/http"
	"strings"

	. "github.com/Phantomvv1/KayTrade/internal/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error only authorized users can access this resource"})
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")
	id, accountType, email, err := ValidateJWT(token)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.Set("id", id)
	c.Set("accountType", accountType)
	c.Set("email", email)

	c.Next()
}
