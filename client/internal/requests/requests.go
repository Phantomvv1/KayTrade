package requests

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	basemodel "github.com/Phantomvv1/KayTrade/client/internal/base_model"
)

var ErrorTokenExpired = errors.New("Error token has expired")

const BaseURL = "http://localhost:42069"

func MakeRequest(method string, urlString string, reader io.Reader, client *http.Client, TokenStore *basemodel.TokenStore) ([]byte, error) {
	req, err := http.NewRequest(method, urlString, reader)
	if err != nil {
		return nil, err
	}

	if TokenStore.Token != "" {
		req.Header.Add("Authorization", "Bearer "+TokenStore.Token)
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

	if res.StatusCode/100 != 2 {
		var info map[string]string
		json.Unmarshal(body, &info)
		if info["error"] == ErrorTokenExpired.Error() {
			log.Println("Refreshing")

			body, err := MakeRequest(http.MethodPost, "http://localhost:42069/refresh", nil, client, TokenStore)
			if err != nil {
				return nil, err
			}

			json.Unmarshal(body, &info)

			TokenStore.Token = info["token"]

			u, err := url.Parse("http://localhost:42069")
			if err != nil {
				return nil, err
			}

			cookies := client.Jar.Cookies(u)
			client.Jar.SetCookies(u, []*http.Cookie{cookies[len(cookies)-1]})

			return MakeRequest(method, urlString, reader, client, TokenStore)
		}

		return nil, errors.New(info["error"])
	}

	return body, nil
}
