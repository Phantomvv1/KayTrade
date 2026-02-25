package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Phantomvv1/KayTrade/internal/auth"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func createTestContext(method, url string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	c.Request = req

	return c, w
}

func TestAdminOnlyMiddleware_Allowed(t *testing.T) {
	c, w := createTestContext("GET", "/", nil)

	c.Set("accountType", auth.Admin)

	AdminOnlyMiddleware(c)

	if w.Code != 0 {
		t.Fatalf("expected no abort, got status %d", w.Code)
	}
}

func TestAdminOnlyMiddleware_Forbidden(t *testing.T) {
	c, w := createTestContext("GET", "/", nil)

	c.Set("accountType", byte(0))

	AdminOnlyMiddleware(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestJSONParserMiddleware_Valid(t *testing.T) {
	body := []byte(`{
		"id": 10,
		"email": "test@mail.com",
		"name": "john"
	}`)

	c, w := createTestContext("POST", "/", body)

	JSONParserMiddleware(c)

	if w.Code != 0 {
		t.Fatalf("unexpected abort")
	}

	if v, ok := c.Get("json_id"); !ok || v.(float64) != 10 {
		t.Fatalf("json_id not set correctly")
	}

	if v, ok := c.Get("name"); !ok || v.(string) != "john" {
		t.Fatalf("name not set correctly")
	}
}

func TestJSONParserMiddleware_InvalidJSON(t *testing.T) {
	body := []byte(`invalid-json`)

	c, w := createTestContext("POST", "/", body)

	JSONParserMiddleware(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSymbolsParserMiddleware_Valid(t *testing.T) {
	c, w := createTestContext("GET", "/?symbols=AAPL&symbols=TSLA", nil)

	SymbolsParserMiddleware(c)

	if w.Code != 0 {
		t.Fatalf("unexpected abort")
	}

	v, ok := c.Get("symbols")
	if !ok || v.(string) != "AAPL,TSLA" {
		t.Fatalf("symbols not joined correctly")
	}
}

func TestSymbolsParserMiddleware_Invalid(t *testing.T) {
	c, w := createTestContext("GET", "/?symbols=", nil)

	SymbolsParserMiddleware(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestStartParserMiddleware_WithStart(t *testing.T) {
	c, _ := createTestContext("GET", "/?start=2024-01-01T00:00:00Z", nil)

	StartParserMiddleware(c)

	v, ok := c.Get("start")
	if !ok || v.(string) != "&start=2024-01-01T00:00:00Z" {
		t.Fatalf("start not set correctly")
	}
}

func TestStartParserMiddleware_DefaultStart(t *testing.T) {
	c, _ := createTestContext("GET", "/", nil)

	StartParserMiddleware(c)

	v, ok := c.Get("start")
	if !ok {
		t.Fatalf("start not set")
	}

	if len(v.(string)) == 0 {
		t.Fatalf("start empty")
	}
}

func resetRateLimiter() {
	mu.Lock()
	defer mu.Unlock()
	rateLimitMap = make(map[string]*rate.Limiter)
}

func TestRateLimiterMiddleware_Allowed(t *testing.T) {
	resetRateLimiter()

	c, w := createTestContext("GET", "/", nil)
	c.Request.RemoteAddr = "127.0.0.1:1234"

	RateLimiterMiddleware(c)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected abort")
	}
}

func TestRateLimiterMiddleware_TooManyRequests(t *testing.T) {
	resetRateLimiter()

	c, w := createTestContext("GET", "/", nil)
	c.Request.RemoteAddr = "127.0.0.1:1234"

	for range 35 {
		RateLimiterMiddleware(c)
	}

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	c, w := createTestContext("GET", "/", nil)

	AuthMiddleware(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
