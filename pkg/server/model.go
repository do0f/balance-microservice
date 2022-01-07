package server

import (
	"balance/pkg/service"
	"errors"

	echo "github.com/labstack/echo/v4"
)

type Balancer interface {
	GetBalance(id int64, currency string) (*service.Balance, error)
	GetHistory(id int64) ([]*service.Transfer, error)
	ChangeBalance(id int64, amount int64) (*service.Balance, error)
	Transfer(senderId int64, recipientId int64, amount int64) (*service.Balance, error)
}

type BalanceHandler struct {
	service Balancer
}

type Server struct {
	*echo.Echo
	handler BalanceHandler
}

var (
	errInvalidParameters = errors.New("invalid request parameters")
)

type errorResponse struct {
	Message string `json:"message"`
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
