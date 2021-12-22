package server

import (
	"balance_microservice/database"
	"fmt"
	"net/http"
	"strconv"

	echo "github.com/labstack/echo/v4"
)

var server *echo.Echo

func Start() error {
	server = echo.New()

	server.GET("/:id/balance", getBalance)
	server.Logger.Fatal(server.Start(":1323"))

	return nil
}

//GET <user id>/balance
func getBalance(context echo.Context) error {
	idString := context.Param("id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return context.String(http.StatusBadRequest, "Invalid id\n")
	}

	balance, err := database.GetBalance(id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return context.String(http.StatusNotFound, "User with such id doesn't exist\n")
		}

		return context.String(http.StatusInternalServerError, "Internal server error\n")
	}

	return context.String(http.StatusOK, fmt.Sprintln(balance))
}
