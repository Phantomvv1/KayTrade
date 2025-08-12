package exit

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorExit(c *gin.Context, status int, message string, err error) {
	if err != nil {
		log.Println(err)
	}
	c.JSON(status, gin.H{"error": "Error " + message})
}

func RequestExit(c *gin.Context, body any, err error, errMsg string) {
	if err.Error() == "Unkown error" {
		log.Println(err)
		c.JSON(http.StatusFailedDependency, body)
		return
	}

	ErrorExit(c, http.StatusFailedDependency, errMsg, err)
}
