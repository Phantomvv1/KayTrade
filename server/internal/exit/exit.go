package exit

import (
	"log"

	"github.com/gin-gonic/gin"
)

func ErrorExit(c *gin.Context, status int, message string, err error) {
	if err != nil {
		log.Println(err)
	}
	c.JSON(status, gin.H{"error": "Error " + message})
}
