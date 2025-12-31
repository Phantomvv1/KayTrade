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

func AuthMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Error only authorized users can access this resource"})
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")
	id, accountType, email, err := ValidateJWT(token)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.Set("id", id)
	c.Set("accountType", accountType)
	c.Set("email", email)

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
		switch k {
		case "id":
			c.Set("json_id", v)
		case "accountType":
			c.Set("json_accountType", v)
		case "email":
			c.Set("json_email", v)
		case "json_id", "json_accountType", "json_email":
			continue
		}
		c.Set(k, v)
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
