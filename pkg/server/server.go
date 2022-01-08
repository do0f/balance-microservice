package server

import (
	"balance/pkg/service"
	"fmt"

	echo "github.com/labstack/echo/v4"
)

const (
	UserBalancePath         string = "balance/users/:id"
	UserBalanceHistoryPath  string = "balance/users/:id/history"
	UserBalanceTransferPath string = "balance/users/:id/transfer"
)

type BalanceService interface {
	GetBalance(id int64, currency string) (*service.Balance, error)
	GetHistory(id int64) ([]*service.Transfer, error)
	ChangeBalance(id int64, amount int64) (*service.Balance, error)
	Transfer(senderId int64, recipientId int64, amount int64) (*service.Balance, error)
}

type Server struct {
	*echo.Echo
	service BalanceService
}

func New(service BalanceService) Server {
	server := Server{echo.New(), service}

	server.GET(UserBalancePath, server.getBalance)
	server.GET(UserBalanceHistoryPath, server.getHistory)
	server.PUT(UserBalancePath, server.changeBalance)
	server.PUT(UserBalanceTransferPath, server.transfer)

	return server
}

func (server *Server) Start(port int) error {
	server.Logger.Fatal(server.Echo.Start(fmt.Sprintf(":%d", port)))
	return nil
}
