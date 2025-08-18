package watchlist

import (
	"net/http"

	. "github.com/Phantomvv1/KayTrade/internal/exit"
	. "github.com/Phantomvv1/KayTrade/internal/requests"
	"github.com/gin-gonic/gin"
)

func CreateWatchlist(c *gin.Context) {
	id := c.Param("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodPost, BaseURL+Trading+id+Watchlist, c.Request.Body, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't create a watchlist for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func GetWatchlist(c *gin.Context) {
	id := c.Param("id")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+Watchlist, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get all the watchlists for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func ManageWatchlist(c *gin.Context) {
	id := c.Param("id")
	watchlistID := c.Param("watchlistId")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodGet, BaseURL+Trading+id+Watchlist+watchlistID, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't get the watchlist for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func UpdateWatchlist(c *gin.Context) {
	id := c.Param("id")
	watchlistID := c.Param("watchlistId")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodPut, BaseURL+Trading+id+Watchlist+watchlistID, c.Request.Body, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't update the watchlist for this account")
		return
	}

	c.JSON(http.StatusOK, body)
}

func DeleteWatchlist(c *gin.Context) {
	id := c.Param("id")
	watchlistID := c.Param("watchlistId")

	headers := BasicAuth()

	body, err := SendRequest[any](http.MethodDelete, BaseURL+Trading+id+Watchlist+watchlistID, nil, nil, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't delete the watchlist")
		return
	}

	c.JSON(http.StatusOK, body)
}

func AddAssetWatchlist(c *gin.Context) {
	id := c.Param("id")
	watchlistID := c.Param("watchlistId")

	headers := BasicAuth()

	errs := map[int]string{
		404: "The requested watchlist is not found, or one of the symbols is not found in the assets",
		422: "Some parameters are not valid",
	}

	body, err := SendRequest[any](http.MethodPost, BaseURL+Trading+id+Watchlist+watchlistID, c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't add an asset to the watchlist")
		return
	}

	c.JSON(http.StatusOK, body)
}

func RemoveSymbolFromWatchlist(c *gin.Context) {
	id := c.Param("id")
	watchlistID := c.Param("watchlistId")
	symbol := c.Param("symbol")

	headers := BasicAuth()

	errs := map[int]string{
		404: "The requested watchlist is not found",
	}

	body, err := SendRequest[any](http.MethodDelete, BaseURL+Trading+id+Watchlist+watchlistID+symbol, c.Request.Body, errs, headers)
	if err != nil {
		RequestExit(c, body, err, "coludn't remove symbol from the watchlist")
		return
	}

	c.JSON(http.StatusOK, body)
}
