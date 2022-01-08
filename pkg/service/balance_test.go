package service_test

import (
	"balance/pkg/repository"
	"balance/pkg/service"
	mock_repository "balance/pkg/service/mocks"
	"errors"
	"math"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBalance_GetBalance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepository := mock_repository.NewMockRepository(mockCtrl)
	mockRepository.EXPECT().BeginTransaction().Return(&repository.Transaction{}, nil).AnyTimes()
	mockRepository.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
	mockRepository.EXPECT().Rollback(gomock.Any()).Return(nil).AnyTimes()

	//default currency
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(1), false).Return(repository.Balance{Amount: 100}, nil).Times(1)
	//usd currency
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(2), false).Return(repository.Balance{Amount: 100}, nil).Times(1)
	//invalid currency
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(3), false).Return(repository.Balance{Amount: 100}, nil).Times(1)
	//non existing user
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(4), false).Return(repository.Balance{}, errors.New("sql: no rows in result set")).Times(1)
	//balance overflow when converting
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(5), false).Return(repository.Balance{Amount: math.MaxInt64 - 1}, nil).Times(1)
	//internal error
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(6), false).Return(repository.Balance{}, errors.New("any error")).Times(1)

	svc := service.New(mockRepository)

	type input struct {
		id       int64
		currency string
	}
	var tests = []struct {
		name        string
		testInput   input
		expectedErr error
	}{
		{name: "default currency", testInput: input{id: 1, currency: "RUB"}, expectedErr: nil},
		{name: "usd currency", testInput: input{id: 2, currency: "USD"}, expectedErr: nil},
		{name: "invalid currency", testInput: input{id: 3, currency: "123"}, expectedErr: service.ErrConvertCurrency},
		{name: "non existing user", testInput: input{id: 4, currency: "RUB"}, expectedErr: service.ErrUserNotFound},
		{name: "balance overflow when converting", testInput: input{id: 5, currency: "VND"}, expectedErr: service.ErrBalanceOverflow},
		{name: "internal error", testInput: input{id: 6, currency: "RUB"}, expectedErr: service.ErrAccessDatabase},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := svc.GetBalance(test.testInput.id, test.testInput.currency)

			//some test may fail because of external API (too many requests for free api key)
			//if so, wait until API will be accessible again to test
			if err != test.expectedErr && err == service.ErrConvertCurrency {
				assert.Fail(t, "External API failure, error converting currency")
			}

			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestBalance_GetHistory(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepository := mock_repository.NewMockRepository(mockCtrl)
	mockRepository.EXPECT().BeginTransaction().Return(&repository.Transaction{}, nil).AnyTimes()
	mockRepository.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
	mockRepository.EXPECT().Rollback(gomock.Any()).Return(nil).AnyTimes()

	//regular query
	mockRepository.EXPECT().GetUserHistory(gomock.Any(), int64(1)).Return([]repository.Transfer{}, nil).Times(1)
	//non existing user
	mockRepository.EXPECT().GetUserHistory(gomock.Any(), int64(2)).Return([]repository.Transfer{}, errors.New("sql: no rows in result set")).Times(1)
	//internal error
	mockRepository.EXPECT().GetUserHistory(gomock.Any(), int64(3)).Return([]repository.Transfer{}, errors.New("any error")).Times(1)

	svc := service.New(mockRepository)

	type input struct {
		id int64
	}
	var tests = []struct {
		name        string
		testInput   input
		expectedErr error
	}{
		{name: "regular query", testInput: input{id: 1}, expectedErr: nil},
		{name: "non existing user", testInput: input{id: 2}, expectedErr: service.ErrUserNotFound},
		{name: "internal error", testInput: input{id: 3}, expectedErr: service.ErrAccessDatabase},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := svc.GetHistory(test.testInput.id)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestBalance_ChangeBalance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepository := mock_repository.NewMockRepository(mockCtrl)
	mockRepository.EXPECT().BeginTransaction().Return(&repository.Transaction{}, nil).AnyTimes()
	mockRepository.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
	mockRepository.EXPECT().Rollback(gomock.Any()).Return(nil).AnyTimes()
	mockRepository.EXPECT().UpdateHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	//add money to balance
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(1), false).Return(repository.Balance{Amount: 200}, nil).Times(1).After(
		mockRepository.EXPECT().ChangeUserBalance(gomock.Any(), int64(1), int64(100)).Return(nil).Times(1).After(
			mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(1), true).Return(repository.Balance{Amount: 100}, nil).Times(1)))
	//get money from balance
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(2), false).Return(repository.Balance{Amount: 0}, nil).Times(1).After(
		mockRepository.EXPECT().ChangeUserBalance(gomock.Any(), int64(2), int64(-100)).Return(nil).Times(1).After(
			mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(2), true).Return(repository.Balance{Amount: 100}, nil).Times(1)))
	//create account
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(3), false).Return(repository.Balance{Amount: 100}, nil).Times(1).After(
		mockRepository.EXPECT().ChangeUserBalance(gomock.Any(), int64(3), int64(100)).Return(nil).Times(1).After(
			mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(3), true).Return(repository.Balance{Amount: 0}, nil).Times(1).After(
				mockRepository.EXPECT().CreateUser(gomock.Any(), int64(3)).Return(nil).Times(1).After(
					mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(3), true).Return(repository.Balance{}, errors.New("sql: no rows in result set")).Times(1)))))
	//try to create with negative amount
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(4), true).Return(repository.Balance{}, errors.New("sql: no rows in result set")).Times(1)
	//trying to withdraw more than account has
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(5), true).Return(repository.Balance{Amount: 0}, nil).Times(1)
	//trying to add too much money
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(6), true).Return(repository.Balance{Amount: math.MaxInt64 - 1}, nil).Times(1)
	//internal error
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(7), true).Return(repository.Balance{}, errors.New("any error")).Times(1)

	svc := service.New(mockRepository)

	type input struct {
		id     int64
		amount int64
	}
	type output struct {
		balance *service.Balance
		err     error
	}
	var tests = []struct {
		name           string
		testInput      input
		expectedOutput output
	}{
		{name: "add money to balance", testInput: input{id: 1, amount: 100}, expectedOutput: output{&service.Balance{2, 0, "RUB"}, nil}},
		{name: "get money from balance", testInput: input{id: 2, amount: -100}, expectedOutput: output{&service.Balance{0, 0, "RUB"}, nil}},
		{name: "create account", testInput: input{id: 3, amount: 100}, expectedOutput: output{&service.Balance{1, 0, "RUB"}, nil}},

		{name: "try to create with negative amount", testInput: input{id: 4, amount: -100}, expectedOutput: output{nil, service.ErrCreatingWithNegativeAmount}},
		{name: "trying to withdraw more than account has", testInput: input{id: 5, amount: -100}, expectedOutput: output{nil, service.ErrNotEnoughMoney}},
		{name: "trying to add too much money", testInput: input{id: 6, amount: 100}, expectedOutput: output{nil, service.ErrBalanceOverflow}},
		{name: "internal error", testInput: input{id: 7, amount: 100}, expectedOutput: output{nil, service.ErrAccessDatabase}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			balance, err := svc.ChangeBalance(test.testInput.id, test.testInput.amount)
			assert.Equal(t, test.expectedOutput.err, err)
			if err == nil {
				assert.Equal(t, test.expectedOutput.balance, balance)
			}
		})
	}
}

func TestBalance_Transfer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepository := mock_repository.NewMockRepository(mockCtrl)
	mockRepository.EXPECT().BeginTransaction().Return(&repository.Transaction{}, nil).AnyTimes()
	mockRepository.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
	mockRepository.EXPECT().Rollback(gomock.Any()).Return(nil).AnyTimes()
	mockRepository.EXPECT().UpdateHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	//regular transfer
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(1), false).Return(repository.Balance{Amount: 0}, nil).Times(1).After(
		mockRepository.EXPECT().ChangeUserBalance(gomock.Any(), int64(1), int64(-100)).Return(nil).Times(1).After(
			mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(1), true).Return(repository.Balance{Amount: 100}, nil).Times(1)))

	mockRepository.EXPECT().ChangeUserBalance(gomock.Any(), int64(2), int64(100)).Return(nil).Times(1).After(
		mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(2), true).Return(repository.Balance{Amount: 0}, nil).Times(1))
	//transfer to non existing user
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(3), true).Return(repository.Balance{Amount: 100}, nil).Times(1)
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(4), true).Return(repository.Balance{}, errors.New("sql: no rows in result set")).Times(1)

	//trying to withdraw more than account has
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(5), true).Return(repository.Balance{Amount: 0}, nil).Times(1)

	//trying to add too much money
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(7), true).Return(repository.Balance{Amount: 100}, nil).Times(1)
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(8), true).Return(repository.Balance{Amount: math.MaxInt64 - 1}, nil).Times(1)

	//internal error
	mockRepository.EXPECT().GetUserBalance(gomock.Any(), int64(9), true).Return(repository.Balance{}, errors.New("any error")).Times(1)

	svc := service.New(mockRepository)

	type input struct {
		senderId    int64
		recipientId int64
		amount      int64
	}
	type output struct {
		balance *service.Balance
		err     error
	}
	var tests = []struct {
		name           string
		testInput      input
		expectedOutput output
	}{
		{name: "regular transfer", testInput: input{senderId: 1, recipientId: 2, amount: 100}, expectedOutput: output{&service.Balance{0, 0, "RUB"}, nil}},

		{name: "transfer to non existing user", testInput: input{senderId: 3, recipientId: 4, amount: 100}, expectedOutput: output{nil, service.ErrUserNotFound}},
		{name: "trying to withdraw more than account has", testInput: input{senderId: 5, recipientId: 6, amount: 100}, expectedOutput: output{nil, service.ErrNotEnoughMoney}},
		{name: "trying to add too much money", testInput: input{senderId: 7, recipientId: 8, amount: 100}, expectedOutput: output{nil, service.ErrBalanceOverflow}},

		{name: "internal error", testInput: input{senderId: 9, recipientId: 10, amount: 100}, expectedOutput: output{nil, service.ErrAccessDatabase}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			balance, err := svc.Transfer(test.testInput.senderId, test.testInput.recipientId, test.testInput.amount)
			assert.Equal(t, test.expectedOutput.err, err)
			if err == nil {
				assert.Equal(t, test.expectedOutput.balance, balance)
			}
		})
	}
}
