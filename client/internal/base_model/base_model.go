package basemodel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Phantomvv1/KayTrade/internal/requests"
)

type BaseModel struct {
	Width, Height int
	Client        *http.Client
	Token         string
}

func (b BaseModel) Refresh() (string, error) {
	reader := strings.NewReader(fmt.Sprintf("{\"token\": \"%s\"}", b.Token))
	body, err := requests.MakeRequest(http.MethodPost, "http://localhost:42069/refresh", reader, b.Client)
	if err != nil {
		return "", err
	}

	var info map[string]string
	err = json.Unmarshal(body, &info)
	if err != nil {
		return "", err
	}

	return info["token"], nil
}
