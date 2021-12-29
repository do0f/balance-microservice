package server

import (
	echo "github.com/labstack/echo/v4"
)

var server *echo.Echo

func Start() error {
	server = echo.New()

	server.GET("balance/users/:id", getBalanceHandler)
	server.GET("balance/users/:id/history", getHistoryHandler)
	server.PUT("balance/users/:id", changeBalanceHandler)
	server.PUT("balance/users/:id/transfer", transferHandler)
	server.Logger.Fatal(server.Start(":1323"))

	return nil
}
