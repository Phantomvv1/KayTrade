package requests

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrorTokenExpired = errors.New("Error token has expired")
var ErrorUnexpected = errors.New("Unexpected error")

func MakeRequest(method string, url string, reader io.Reader, client *http.Client) ([]byte, error) {
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
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
		return body, ErrorUnexpected
	}

	return body, nil
}
