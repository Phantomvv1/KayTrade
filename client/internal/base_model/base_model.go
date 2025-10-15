package basemodel

import (
	"net/http"
)

type BaseModel struct {
	Width, Height int
	Change        bool
	Client        *http.Client
	Token         string
}
