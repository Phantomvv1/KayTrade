package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter()
}

func performRequest(r http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRouter_Root(t *testing.T) {
	r := setupRouter()
	w := performRequest(r, http.MethodGet, "/", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRouter_SignUpExists(t *testing.T) {
	r := setupRouter()
	w := performRequest(r, http.MethodPost, "/sign-up", nil)

	if w.Code == http.StatusNotFound {
		t.Fatal("POST /sign-up route not registered")
	}
}

func TestRouter_LoginExists(t *testing.T) {
	r := setupRouter()
	w := performRequest(r, http.MethodPost, "/log-in", nil)

	if w.Code == http.StatusNotFound {
		t.Fatal("POST /log-in route not registered")
	}
}

func TestProtectedRoute_WithoutAuth(t *testing.T) {
	r := setupRouter()
	w := performRequest(r, http.MethodGet, "/users", nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAdminRoute_WithoutAuth(t *testing.T) {
	r := setupRouter()
	w := performRequest(r, http.MethodGet, "/users/all", nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRateLimiterIsApplied(t *testing.T) {
	r := setupRouter()

	var lastCode int
	for range 40 {
		w := performRequest(r, http.MethodGet, "/", nil)
		lastCode = w.Code
	}

	if lastCode != http.StatusTooManyRequests {
		t.Fatalf("expected rate limit (429), got %d", lastCode)
	}
}

func TestMarketDataRoutesExist(t *testing.T) {
	r := setupRouter()

	paths := []string{
		"/data/auctions?symbols=AAPL",
		"/data/bars?symbols=AAPL",
		"/data/bars/latest?symbols=AAPL",
		"/data/exchanges",
		"/data/stocks/most-active",
		"/data/stocks/top-market-movers",
	}

	for _, path := range paths {
		w := performRequest(r, http.MethodGet, path, nil)
		if w.Code == http.StatusNotFound {
			t.Fatalf("route %s not registered", path)
		}
	}
}

func TestWebSocketRouteExists(t *testing.T) {
	r := setupRouter()
	w := performRequest(r, http.MethodGet, "/data/stocks/live/AAPL", nil)

	// Will fail because it's not a WS upgrade — but route exists
	if w.Code == http.StatusNotFound {
		t.Fatal("websocket route not registered")
	}
}
