package middleware

import (
	"encoding/json"
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
	id, accountType, email, err := ValidateJWT(token, false)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.Set("id", id)
	c.Set("accountType", accountType)
	c.Set("email", email)

	c.Next()
}

func ParserMiddleware(c *gin.Context) {
	var information map[string]any
	err := json.NewDecoder(c.Request.Body).Decode(&information)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error unable to parse the body of the request"})
		c.Abort()
		return
	}

	for k, v := range information {
		c.Set(k, v)
	}

	c.Next()
}
