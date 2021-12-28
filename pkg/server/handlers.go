package server

import (
	"balance_microservice/balance"
	"net/http"

	echo "github.com/labstack/echo/v4"
)

//GET balance/users/<user id>
//returns error and balance struct in JSON
func getBalanceHandler(context echo.Context) error {
	request := &getData{}
	err := context.Bind(request)
	if err != nil || request.Id < 0 {
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

//PUT balance/users/<user id>
//JSON: amount: <amount of kopecks>
//returns error and changed balance struct in JSON
func changeBalanceHandler(context echo.Context) error {
	request := &changeData{}
	err := context.Bind(request)
	if err != nil || request.Id < 0 {
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

//PUT balance/users/<user id>/transfer
//JSON amount: <amount of kopeks> recipient: <recipient's id>
//returns error and changed balance struct of sender in JSON
func transferHandler(context echo.Context) error {
	request := &transferData{}
	err := context.Bind(request)
	//check for self-transfer, negative transfer and invalid ids
	if err != nil || request.SenderId < 0 || request.RecipientId < 0 || request.SenderId == request.RecipientId || request.Amount <= 0 {
		return context.JSON(http.StatusBadRequest, newResponse(nil, errInvalidParameters))
	}

	balanceStruct, err := balance.TransferTransaction(request.SenderId, request.RecipientId, request.Amount)
	switch err {
	case nil:
		return context.JSON(http.StatusOK, newResponse(balanceStruct, nil))

	case balance.ErrNotEnoughMoney:
		fallthrough
	case balance.ErrUserNotFound:
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
