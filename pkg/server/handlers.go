package server

import (
	"balance/pkg/service"
	"errors"
	"net/http"

	echo "github.com/labstack/echo/v4"
)

var (
	errInvalidParameters = errors.New("invalid request parameters")
)

//structures for getting data from requests and sending responses
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

//GET balance/users/<user id>?currency=<currency name>
//returns error and balance struct in JSON
func (s *Server) getBalance(ctx echo.Context) error {
	request := &getData{}
	err := ctx.Bind(request)
	if err != nil || request.Id < 0 {
		return ctx.JSON(http.StatusBadRequest, errorResponse{errInvalidParameters.Error()})
	}

	if request.Currency == "" {
		request.Currency = "RUB"
	}

	balanceStruct, err := s.service.GetBalance(request.Id, request.Currency)
	if err == nil {
		return ctx.JSON(http.StatusOK, balanceStruct)
	}

	var code int
	switch err {
	case service.ErrUserNotFound:
		code = http.StatusNotFound
	case service.ErrConvertCurrency:
		code = http.StatusBadRequest

	default:
		code = http.StatusInternalServerError
	}
	return ctx.JSON(code, errorResponse{err.Error()})
}

//GET balance/users/<user id>/history
//returns error and user's transfers in JSON
func (s *Server) getHistory(ctx echo.Context) error {
	request := &getData{}
	err := ctx.Bind(request)
	if err != nil || request.Id < 0 {
		return ctx.JSON(http.StatusBadRequest, errorResponse{errInvalidParameters.Error()})
	}

	transfers, err := s.service.GetHistory(request.Id)

	if err == nil {
		return ctx.JSON(http.StatusOK, transfers)
	}

	var code int
	switch err {
	case service.ErrUserNotFound:
		code = http.StatusNotFound

	default:
		code = http.StatusInternalServerError
	}
	return ctx.JSON(code, errorResponse{err.Error()})
}

//PUT balance/users/<user id>
//JSON: amount: <amount of kopecks>
//returns error and changed balance struct in JSON
func (s *Server) changeBalance(ctx echo.Context) error {
	request := &changeData{}
	err := ctx.Bind(request)
	if err != nil || request.Id < 0 {
		return ctx.JSON(http.StatusBadRequest, errorResponse{errInvalidParameters.Error()})
	}

	balanceStruct, err := s.service.ChangeBalance(request.Id, request.Amount)

	if err == nil {
		return ctx.JSON(http.StatusOK, balanceStruct)
	}

	var code int
	switch err {
	case service.ErrNotEnoughMoney:
		fallthrough
	case service.ErrCreatingWithNegativeAmount:
		code = http.StatusBadRequest

	default:
		code = http.StatusInternalServerError
	}
	return ctx.JSON(code, errorResponse{err.Error()})
}

//PUT balance/users/<user id>/transfer
//JSON amount: <amount of kopeks> recipient: <recipient's id>
//returns error and changed balance struct of sender in JSON
func (s *Server) transfer(ctx echo.Context) error {
	request := &transferData{}
	err := ctx.Bind(request)
	//check for self-transfer, negative transfer and invalid ids
	if err != nil || request.SenderId < 0 || request.RecipientId < 0 || request.SenderId == request.RecipientId || request.Amount <= 0 {
		return ctx.JSON(http.StatusBadRequest, errorResponse{errInvalidParameters.Error()})
	}

	balanceStruct, err := s.service.Transfer(request.SenderId, request.RecipientId, request.Amount)

	if err == nil {
		return ctx.JSON(http.StatusOK, balanceStruct)
	}

	var code int
	switch err {
	case service.ErrNotEnoughMoney:
		fallthrough
	case service.ErrUserNotFound:
		code = http.StatusBadRequest

	default:
		code = http.StatusInternalServerError
	}
	return ctx.JSON(code, errorResponse{err.Error()})
}
