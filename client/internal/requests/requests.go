package requests

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrorTokenExpired = errors.New("Error token has expired")

const BaseURL = "http://localhost:42069"

func MakeRequest(method string, url string, reader io.Reader, client *http.Client, token string) ([]byte, error) {
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	errMsg := fmt.Sprintf("{\"%s\": \"%s\"}", "error", "Error only authorized users can access this resource")
	if string(body) == errMsg {
		return nil, ErrorTokenExpired
	}

	if res.StatusCode/100 != 2 {
		var info map[string]string
		json.Unmarshal(body, &info)
		return nil, errors.New(info["error"])
	}

	return body, nil
}
