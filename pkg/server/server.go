package server

import (
	"balance_microservice/balance"
	"fmt"
	"net/http"

	echo "github.com/labstack/echo/v4"
)

var server *echo.Echo

var (
	errInvalidParameters = fmt.Errorf("invalid request parameters")
)

func Start() error {
	server = echo.New()

	server.GET("users/:id/balance", getBalanceHandler)
	server.PUT("users/:id/balance", changeBalanceHandler)
	server.Logger.Fatal(server.Start(":1323"))

	return nil
}

//GET <user id>/balance
func getBalanceHandler(context echo.Context) error {
	request := &getData{}
	err := context.Bind(request)
	if request.Id < 0 || err != nil {
		return context.JSON(http.StatusBadRequest, newResponse(nil, errInvalidParameters))
	}

	balanceStruct, err := balance.GetBalanceTransaction(request.Id)

	switch err {
	case nil:
		return context.JSON(http.StatusOK, newResponse(balanceStruct, err))
	case balance.ErrUserNotFound:
		return context.JSON(http.StatusNotFound, newResponse(nil, err))
	case balance.ErrAccessDatabase:
		fallthrough
	case balance.ErrNegativeBalance:
		fallthrough
	default:
		return context.JSON(http.StatusInternalServerError, newResponse(nil, err))
	}

}

//PUT <user id>/balance>
//JSON: amount: <amount of kopecks>
func changeBalanceHandler(context echo.Context) error {
	request := &changeData{}
	err := context.Bind(request)
	if request.Id < 0 || err != nil {
		return context.JSON(http.StatusBadRequest, newResponse(nil, errInvalidParameters))
	}

	balanceStruct, err := balance.ChangeBalanceTransaction(request.Id, request.Amount)

	switch err {
	case nil:
		return context.JSON(http.StatusOK, newResponse(balanceStruct, nil))

	case balance.ErrNotEnoughMoney:
		fallthrough
	case balance.ErrCreatingWithNegativeAmount:
		return context.JSON(http.StatusBadRequest, newResponse(nil, err))

	case balance.ErrAccessDatabase:
		fallthrough
	case balance.ErrNegativeBalance:
		fallthrough
	case balance.ErrBalanceOverflow:
		fallthrough
	default:
		return context.JSON(http.StatusInternalServerError, newResponse(nil, err))
	}

}

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

func newResponse(record *balance.Balance, err error) *response {
	if err == nil {
		return &response{nil, record}
	}

	errString := err.Error()
	return &response{&errString, record}
}
