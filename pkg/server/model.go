package server

import (
	"fmt"
)

var (
	errInvalidParameters = fmt.Errorf("invalid request parameters")
)

type response struct {
	Err    *string     `json:"error"`
	Record interface{} `json:"data"`
}
type getData struct {
	Id       int64  `param:"id"`
	Currency string `query:"currency"`
}
type changeData struct {
	Id     int64 `param:"id"`
	Amount int64 `json:"amount"`
}
type transferData struct {
	SenderId    int64 `param:"id"`
	RecipientId int64 `json:"recipient"`
	Amount      int64 `json:"amount"`
}

func newResponse(record interface{}, err error) *response {
	if err == nil {
		return &response{nil, record}
	}

	errString := err.Error()
	return &response{&errString, record}
}
