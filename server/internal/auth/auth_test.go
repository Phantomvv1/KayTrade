package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestSHA512(t *testing.T) {
	got := SHA512("test")
	want := "ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f8819a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc5fa9ad8e6f57f50028a8ff"

	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestGenerateAndValidateJWT(t *testing.T) {
	t.Setenv("JWT_KEY", "test-secret")

	token, err := GenerateJWT("user-id", User, "test@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	id, accType, email, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if id != "user-id" {
		t.Fatalf("expected id=user-id, got %s", id)
	}

	if accType != User {
		t.Fatalf("expected User, got %d", accType)
	}

	if email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", email)
	}
}

func TestValidateJWTExpired(t *testing.T) {
	t.Setenv("JWT_KEY", "test-secret")

	claims := jwt.MapClaims{
		"id":         "x",
		"type":       User,
		"email":      "a@b.com",
		"expiration": time.Now().Add(-time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("test-secret"))

	_, _, _, err := ValidateJWT(signed)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestLogInUserNotFound(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://invalid")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(`{"email":"x@y.com","password":"123"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	c.Request = req

	LogIn(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
