package basemodel

import (
	"net/http"
)

type BaseModel struct {
	Width, Height int
	Client        *http.Client
	Token         string
}
