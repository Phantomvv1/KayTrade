package exit

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestErrorExit_WithError(t *testing.T) {
	c, w := createTestContext()

	ErrorExit(c, http.StatusBadRequest, "something went wrong", nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "Error something went wrong"
	if resp["error"] != expected {
		t.Fatalf("expected %q, got %q", expected, resp["error"])
	}
}

func TestErrorExit_WithoutError(t *testing.T) {
	c, w := createTestContext()

	ErrorExit(c, http.StatusInternalServerError, "failed", nil)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestRequestExit_UnknownError(t *testing.T) {
	c, w := createTestContext()

	body := gin.H{"test": "value"}
	err := errors.New("Unkown error")

	RequestExit(c, body, err, "ignored")

	if w.Code != http.StatusFailedDependency {
		t.Fatalf("expected 424, got %d", w.Code)
	}

	expected := `{"test":"value"}`
	if w.Body.String() != expected {
		t.Fatalf("expected body %s, got %s", expected, w.Body.String())
	}
}

func TestRequestExit_NilBodyWithError(t *testing.T) {
	c, w := createTestContext()

	err := errors.New("db failed")

	RequestExit(c, nil, err, "ignored")

	if w.Code != http.StatusFailedDependency {
		t.Fatalf("expected 424, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	expected := "Error db failed"
	if resp["error"] != expected {
		t.Fatalf("expected %q, got %q", expected, resp["error"])
	}
}

func TestRequestExit_DefaultCase(t *testing.T) {
	c, w := createTestContext()

	err := errors.New("custom message")

	RequestExit(c, gin.H{"x": 1}, err, "custom message")

	if w.Code != http.StatusFailedDependency {
		t.Fatalf("expected 424, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	expected := "Error custom message"
	if resp["error"] != expected {
		t.Fatalf("expected %q, got %q", expected, resp["error"])
	}
}
