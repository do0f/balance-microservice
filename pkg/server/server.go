package server

import (
	echo "github.com/labstack/echo/v4"
)

func New(service Balancer) Server {
	server := Server{echo.New(), BalanceHandler{service}}

	server.GET("balance/users/:id", server.handler.getBalance)
	server.GET("balance/users/:id/history", server.handler.getHistory)
	server.PUT("balance/users/:id", server.handler.changeBalance)
	server.PUT("balance/users/:id/transfer", server.handler.transfer)

	return server
}

func (server *Server) Start() error {
	server.Logger.Fatal(server.Echo.Start(":1323"))
	return nil
}
