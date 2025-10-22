package messages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	LandingPageNumber = iota
	WatchlistPageNumber
	ErrorPageNumber
)

type PageSwitchMsg struct {
	Page int
	Err  error
}

func Refresh(token string, client *http.Client) (string, error) {
	reader := bytes.NewReader([]byte(fmt.Sprintf("\"token\": \"%s\"", token)))
	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/refresh", reader)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result["token"], nil
}
