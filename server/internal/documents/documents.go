package documents

import (
	"net/http"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
)

func GetAllDocuments(c *gin.Context) {
	id := c.Param("id")

	headers := BasicAuth()

	errs := map[int]string{
		404: "Not found",
	}

	body, err := SendRequest[any](http.MethodGet, BaseURL+Accounts+id+"/"+Documents, nil, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the documents for your account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func DownloadDocument(c *gin.Context) {
	id := c.Param("id")
	documentID := c.Param("documentId")

	headers := BasicAuth()

	errs := map[int]string{
		404: "Document is not found",
	}

	req, err := http.NewRequest(http.MethodGet, BaseURL+Accounts+id+"/"+Documents+documentID, nil)
	if err != nil {
		ErrorExit(c, http.StatusFailedDependency, "couldn't create the request", err)
		return
	}

	for header, value := range headers {
		req.Header.Add(header, value)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		ErrorExit(c, http.StatusFailedDependency, "couldn't make the request", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		for code, errMsg := range errs {
			if res.StatusCode == code {
				ErrorExit(c, http.StatusFailedDependency, errMsg, nil)
				return
			}
		}
	}

	switch res.StatusCode {
	case http.StatusMovedPermanently:
		c.Redirect(http.StatusMovedPermanently, res.Header["Location"][0])
		return

	case http.StatusOK:
		c.JSON(http.StatusOK, res.Body)
		return

	default:
		c.JSON(http.StatusFailedDependency, gin.H{"error": "Error while trying to get the headers"})
	}
}
