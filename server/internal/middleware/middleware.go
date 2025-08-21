package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	. "github.com/Phantomvv1/KayTrade/internal/auth"
	"github.com/gin-gonic/gin"
)

func AuthParserMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Error only authorized users can access this resource"})
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")
	id, accountType, email, err := ValidateJWT(token, false)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.Set("id", id)
	c.Set("accountType", accountType)
	c.Set("token_email", email)

	c.Next()
}

func AdminOnlyMiddleware(c *gin.Context) {
	accType, _ := c.Get("accountType")
	accountType := accType.(byte)
	if accountType != Admin {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Error only admins can access this resource"})
		return
	}

	c.Next()
}

func JSONParserMiddleware(c *gin.Context) {
	var information map[string]any
	err := json.NewDecoder(c.Request.Body).Decode(&information)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Error unable to parse the body of the request"})
		return
	}

	for k, v := range information {
		c.Set(k, v)
	}

	c.Next()
}

func AuthProtectMiddleware(c *gin.Context) {
	userID := c.Param("id")
	if strings.HasPrefix(userID, "/") && strings.HasSuffix(userID, "/") {
		userID = strings.TrimPrefix(userID, "/")
		userID = strings.TrimSuffix(userID, "/")
	}

	id := c.GetString("id")
	acc, _ := c.Get("accountType")
	accountType := acc.(byte)

	if accountType != Admin && id != userID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Error only admins and the user themselves can access this resource"})
		return
	}

	c.Next()
}

func SymbolsParserMiddleware(c *gin.Context) {
	symbols := c.QueryArray("symbols")
	if symbols == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No information given"})
		return
	} else if symbols[0] == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No information given"})
		return
	}

	symbolsToSend := strings.Join(symbols, ",")
	c.Set("symbols", symbolsToSend)

	c.Next()
}

func StartParserMiddleware(c *gin.Context) {
	start := c.Query("start")
	if start == "" {
		start = "&start=" + time.Now().UTC().Truncate(time.Hour*24).Format(time.RFC3339)
	} else {
		start = "&start=" + start
	}

	c.Set("start", start)

	c.Next()
}
