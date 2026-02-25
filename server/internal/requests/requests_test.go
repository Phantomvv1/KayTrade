package requests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSendRequest_SuccessJSON(t *testing.T) {
	type Response struct {
		Message string `json:"message"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("accept") != "application/json" {
			t.Fatal("missing accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Message: "ok"})
	}))
	defer ts.Close()

	res, err := SendRequest[Response](
		http.MethodGet,
		ts.URL,
		nil,
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Message != "ok" {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestSendRequest_ErrorWithMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"bad request"}`))
	}))
	defer ts.Close()

	_, err := SendRequest[any](
		http.MethodGet,
		ts.URL,
		nil,
		nil,
		nil,
	)

	if err == nil || err.Error() != "bad request" {
		t.Fatalf("expected error 'bad request', got %v", err)
	}
}

func TestSendRequest_ErrorFromErrMap(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	errs := map[int]string{
		http.StatusUnauthorized: "unauthorized",
	}

	_, err := SendRequest[any](
		http.MethodGet,
		ts.URL,
		nil,
		errs,
		nil,
	)

	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected error 'unauthorized', got %v", err)
	}
}

func TestSendRequest_TextPlainReturnsUnknownError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("something broke"))
	}))
	defer ts.Close()

	_, err := SendRequest[any](
		http.MethodGet,
		ts.URL,
		nil,
		nil,
		nil,
	)

	if err == nil || err.Error() != "Unknown error" {
		t.Fatalf("expected 'Unkown error', got %v", err)
	}
}

func TestSendRequest_EmptyBodyIsOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	_, err := SendRequest[any](
		http.MethodGet,
		ts.URL,
		nil,
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendRequest_CustomHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "123" {
			t.Fatal("missing custom header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	headers := map[string]string{"X-Test": "123"}

	_, err := SendRequest[any](
		http.MethodPost,
		ts.URL,
		bytes.NewBufferString("{}"),
		nil,
		headers,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBasicAuth(t *testing.T) {
	t.Setenv("API_KEY", "key123")
	t.Setenv("SECRET_KEY", "secret456")

	h := BasicAuth()

	auth, ok := h["Authorization"]
	if !ok {
		t.Fatal("Authorization header missing")
	}

	expected := base64.StdEncoding.EncodeToString([]byte("key123:secret456"))
	if auth != "Basic "+expected {
		t.Fatalf("unexpected auth header: %s", auth)
	}
}

func TestBasicAuth_EmptyEnv(t *testing.T) {
	os.Unsetenv("API_KEY")
	os.Unsetenv("SECRET_KEY")

	h := BasicAuth()
	if h["Authorization"] == "" {
		t.Fatal("Authorization header should still exist")
	}
}
