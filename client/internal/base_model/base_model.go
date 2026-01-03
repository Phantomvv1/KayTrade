package basemodel

import (
	"net/http"
)

type BaseModel struct {
	Width, Height int
	Client        *http.Client
	TokenStore    *TokenStore
}

type TokenStore struct {
	Token string
}
