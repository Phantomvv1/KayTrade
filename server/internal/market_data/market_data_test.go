package marketdata

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTimeFrameValid_ValidCases(t *testing.T) {
	tests := []string{
		"1T", "5T", "59T",
		"1H", "23H",
		"1D", "1W",
		"1M", "2M", "3M", "4M", "6M", "12M",
	}

	for _, tt := range tests {
		tf := TimeFrame(tt)
		if !tf.ValidTimeFrame() {
			t.Fatalf("expected %s to be valid", tt)
		}
	}
}

func TestTimeFrameValid_InvalidCases(t *testing.T) {
	tests := []string{
		"",
		"T", "H", "M",
		"60T", "-1T",
		"24H", "-1H",
		"5Y",
		"5M", "7M", "13M",
		"ABC",
	}

	for _, tt := range tests {
		tf := TimeFrame(tt)
		if tf.ValidTimeFrame() {
			t.Fatalf("expected %s to be invalid", tt)
		}
	}
}

func createGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestGetHistoricalBars_InvalidTimeFrame(t *testing.T) {
	c, w := createGinContext()

	c.Set("symbols", "AAPL")
	c.Set("start", "&start=2024-01-01")
	c.Request = httptest.NewRequest("GET", "/?timeframe=BAD", nil)

	GetHistoricalBars(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetConditionCodes_InvalidTickType(t *testing.T) {
	c, w := createGinContext()

	c.Params = gin.Params{{Key: "ticktype", Value: "bad"}}
	c.Request = httptest.NewRequest("GET", "/?tape=A", nil)

	GetConditionCodes(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetConditionCodes_InvalidTape(t *testing.T) {
	c, w := createGinContext()

	c.Params = gin.Params{{Key: "ticktype", Value: "trade"}}
	c.Request = httptest.NewRequest("GET", "/?tape=Z", nil)

	GetConditionCodes(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetMostActiveStocks_InvalidBy(t *testing.T) {
	c, w := createGinContext()
	c.Request = httptest.NewRequest("GET", "/?by=invalid", nil)

	GetMostActiveStocks(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetMostActiveStocks_InvalidTop(t *testing.T) {
	c, w := createGinContext()
	c.Request = httptest.NewRequest("GET", "/?top=999", nil)

	GetMostActiveStocks(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetTopMarketMovers_InvalidTop(t *testing.T) {
	c, w := createGinContext()
	c.Request = httptest.NewRequest("GET", "/?top=999", nil)

	GetTopMarketMovers(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
