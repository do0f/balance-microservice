package server

import (
	mock_service "balance/pkg/server/mocks"
	"balance/pkg/service"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlers_GetBalance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_service.NewMockBalancer(mockCtrl)

	//default currency & RUB currency explicitly
	mockService.EXPECT().GetBalance(int64(1), "RUB").Return(&service.Balance{}, nil).Times(2)
	//USD currency
	mockService.EXPECT().GetBalance(int64(2), "USD").Return(&service.Balance{}, nil).Times(1)
	//non existing user
	mockService.EXPECT().GetBalance(int64(3), "RUB").Return(nil, service.ErrUserNotFound).Times(1)
	//invalid currency
	mockService.EXPECT().GetBalance(int64(1), "123").Return(nil, service.ErrConvertCurrency).Times(1)
	//internal error
	mockService.EXPECT().GetBalance(int64(4), "RUB").Return(nil, service.ErrAccessDatabase).Times(1)

	server := New(mockService)
	go server.Start(1324)

	type input struct {
		id       string
		currency string
	}
	var tests = []struct {
		name         string
		testInput    input
		expectedCode int
	}{
		{name: "too long id", testInput: input{id: "12345678987654123123415235231324234", currency: ""}, expectedCode: http.StatusBadRequest},
		{name: "negative id", testInput: input{id: "-5", currency: ""}, expectedCode: http.StatusBadRequest},
		{name: "string id", testInput: input{id: "stringid", currency: ""}, expectedCode: http.StatusBadRequest},

		{name: "default currency", testInput: input{id: "1", currency: ""}, expectedCode: http.StatusOK},
		{name: "RUB currency explicitly", testInput: input{id: "1", currency: "RUB"}, expectedCode: http.StatusOK},
		{name: "USD currency", testInput: input{id: "2", currency: "USD"}, expectedCode: http.StatusOK},
		{name: "non existing user", testInput: input{id: "3", currency: "RUB"}, expectedCode: http.StatusNotFound},
		{name: "invalid currency", testInput: input{id: "1", currency: "123"}, expectedCode: http.StatusBadRequest},

		{name: "internal error", testInput: input{id: "4", currency: "RUB"}, expectedCode: http.StatusInternalServerError},
	}

	for _, test := range tests {
		var request *http.Request
		if test.testInput.currency != "" {
			query := make(url.Values)
			query.Set("currency", test.testInput.currency)
			request = httptest.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)
		} else {
			request = httptest.NewRequest(http.MethodGet, "/", nil)
		}

		recorder := httptest.NewRecorder()
		ctx := server.NewContext(request, recorder)
		ctx.SetPath(UserBalancePath)
		ctx.SetParamNames("id")
		ctx.SetParamValues(test.testInput.id)

		t.Run(test.name, func(t *testing.T) {
			if assert.NoError(t, server.getBalance(ctx)) {
				assert.Equal(t, test.expectedCode, recorder.Code)
			}
		})
	}
}

func TestHandlers_GetHistory(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_service.NewMockBalancer(mockCtrl)

	//regular history
	mockService.EXPECT().GetHistory(int64(1)).Return([]*service.Transfer{}, nil).Times(1)
	//non existing user
	mockService.EXPECT().GetHistory(int64(2)).Return(nil, service.ErrUserNotFound).Times(1)
	//internal error
	mockService.EXPECT().GetHistory(int64(3)).Return(nil, service.ErrAccessDatabase).Times(1)

	server := New(mockService)
	go server.Start(1325)

	type input struct {
		id string
	}
	var tests = []struct {
		name         string
		testInput    input
		expectedCode int
	}{
		{name: "too long id", testInput: input{id: "12345678987654123123415235231324234"}, expectedCode: http.StatusBadRequest},
		{name: "negative id", testInput: input{id: "-4"}, expectedCode: http.StatusBadRequest},
		{name: "string id", testInput: input{id: "stringid"}, expectedCode: http.StatusBadRequest},

		{name: "regular history", testInput: input{id: "1"}, expectedCode: http.StatusOK},
		{name: "non existing user", testInput: input{id: "2"}, expectedCode: http.StatusNotFound},

		{name: "internal error", testInput: input{id: "3"}, expectedCode: http.StatusInternalServerError},
	}

	for _, test := range tests {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()
		ctx := server.NewContext(request, recorder)
		ctx.SetPath(UserBalancePath)
		ctx.SetParamNames("id")
		ctx.SetParamValues(test.testInput.id)

		t.Run(test.name, func(t *testing.T) {
			if assert.NoError(t, server.getHistory(ctx)) {
				assert.Equal(t, test.expectedCode, recorder.Code)
			}
		})
	}
}

func TestHandlers_ChangeBalance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_service.NewMockBalancer(mockCtrl)
	//positive value
	mockService.EXPECT().ChangeBalance(int64(1), int64(100)).Return(&service.Balance{}, nil).Times(1)
	//negative value
	mockService.EXPECT().ChangeBalance(int64(1), int64(-100)).Return(&service.Balance{}, nil).Times(1)
	//not enough money
	mockService.EXPECT().ChangeBalance(int64(2), int64(-10000)).Return(nil, service.ErrNotEnoughMoney).Times(1)
	//creating new account with negative amount
	mockService.EXPECT().ChangeBalance(int64(3), int64(-100)).Return(nil, service.ErrCreatingWithNegativeAmount).Times(1)
	//internal error
	mockService.EXPECT().ChangeBalance(int64(4), int64(1)).Return(nil, service.ErrBalanceOverflow).Times(1)

	server := New(mockService)
	go server.Start(1326)

	type input struct {
		id     string
		amount string
	}
	var tests = []struct {
		name         string
		testInput    input
		expectedCode int
	}{
		{name: "positive value", testInput: input{id: "1", amount: "100"}, expectedCode: http.StatusOK},
		{name: "negative value", testInput: input{id: "1", amount: "-100"}, expectedCode: http.StatusOK},
		{name: "not enough money", testInput: input{id: "2", amount: "-10000"}, expectedCode: http.StatusBadRequest},
		{name: "creating new account with negative amount", testInput: input{id: "3", amount: "-100"}, expectedCode: http.StatusBadRequest},

		{name: "too long id", testInput: input{id: "12345678987654123123415235231324234", amount: "1"}, expectedCode: http.StatusBadRequest},
		{name: "negative id", testInput: input{id: "-5", amount: "1"}, expectedCode: http.StatusBadRequest},
		{name: "string id", testInput: input{id: "stringid", amount: "1"}, expectedCode: http.StatusBadRequest},
		{name: "invalid amount", testInput: input{id: "1", amount: "stringamount"}, expectedCode: http.StatusBadRequest},

		{name: "internal error", testInput: input{id: "4", amount: "1"}, expectedCode: http.StatusInternalServerError},
	}

	for _, test := range tests {
		jsonBody := strings.NewReader(fmt.Sprintf(`{ "amount": %s }`, test.testInput.amount))
		request := httptest.NewRequest(http.MethodPut, "/", jsonBody)
		request.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		ctx := server.NewContext(request, recorder)
		ctx.SetPath(UserBalancePath)
		ctx.SetParamNames("id")
		ctx.SetParamValues(test.testInput.id)

		t.Run(test.name, func(t *testing.T) {
			if assert.NoError(t, server.changeBalance(ctx)) {
				assert.Equal(t, test.expectedCode, recorder.Code)
			}
		})
	}
}

func TestHandlers_Transfer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_service.NewMockBalancer(mockCtrl)

	//regular transfer
	mockService.EXPECT().Transfer(int64(1), int64(2), int64(100)).Return(&service.Balance{}, nil).Times(1)
	//not enough money
	mockService.EXPECT().Transfer(int64(1), int64(4), int64(1000000)).Return(nil, service.ErrNotEnoughMoney).Times(1)
	//non existing user
	mockService.EXPECT().Transfer(int64(1), int64(5), int64(100)).Return(nil, service.ErrUserNotFound).Times(1)
	//internal error
	mockService.EXPECT().Transfer(int64(1), int64(6), int64(100)).Return(nil, service.ErrAccessDatabase).Times(1)

	server := New(mockService)
	go server.Start(1327)

	type input struct {
		senderId    string
		recipientId string
		amount      string
	}
	var tests = []struct {
		name         string
		testInput    input
		expectedCode int
	}{
		{name: "regular transfer", testInput: input{senderId: "1", recipientId: "2", amount: "100"}, expectedCode: http.StatusOK},
		{name: "not enough money", testInput: input{senderId: "1", recipientId: "4", amount: "1000000"}, expectedCode: http.StatusBadRequest},
		{name: "non existing user", testInput: input{senderId: "1", recipientId: "5", amount: "100"}, expectedCode: http.StatusBadRequest},

		{name: "too long id", testInput: input{senderId: "12345678987654123123415235231324234", recipientId: "1", amount: "1"}, expectedCode: http.StatusBadRequest},
		{name: "negative id", testInput: input{senderId: "-5", recipientId: "1", amount: "1"}, expectedCode: http.StatusBadRequest},
		{name: "string id", testInput: input{senderId: "stringid", recipientId: "1", amount: "1"}, expectedCode: http.StatusBadRequest},
		{name: "transfering to oneself", testInput: input{senderId: "1", recipientId: "1", amount: "100"}, expectedCode: http.StatusBadRequest},
		{name: "negative amount transfer", testInput: input{senderId: "1", recipientId: "3", amount: "-100"}, expectedCode: http.StatusBadRequest},
		{name: "bad amount", testInput: input{senderId: "1", recipientId: "2", amount: "stringamount"}, expectedCode: http.StatusBadRequest},

		{name: "internal error", testInput: input{senderId: "1", recipientId: "6", amount: "100"}, expectedCode: http.StatusInternalServerError},
	}

	for _, test := range tests {
		jsonBody := strings.NewReader(fmt.Sprintf(`{ "recipient": %s, "amount": %s }`, test.testInput.recipientId, test.testInput.amount))
		request := httptest.NewRequest(http.MethodPut, "/", jsonBody)
		request.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		ctx := server.NewContext(request, recorder)
		ctx.SetPath(UserBalancePath)
		ctx.SetParamNames("id")
		ctx.SetParamValues(test.testInput.senderId)

		t.Run(test.name, func(t *testing.T) {
			if assert.NoError(t, server.transfer(ctx)) {
				assert.Equal(t, test.expectedCode, recorder.Code)
			}
		})
	}
}
