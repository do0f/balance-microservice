package server

import (
	"balance_microservice/balance"
	"fmt"
)

var (
	errInvalidParameters = fmt.Errorf("invalid request parameters")
)

type response struct {
	Err    *string          `json:"error"`
	Record *balance.Balance `json:"balance"`
}
type getData struct {
	Id int64 `param:"id"`
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

func newResponse(record *balance.Balance, err error) *response {
	if err == nil {
		return &response{nil, record}
	}

	errString := err.Error()
	return &response{&errString, record}
}
