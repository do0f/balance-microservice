package service

import (
	"balance/pkg/repository"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/tidwall/gjson"
)

type Repository interface {
	Open() error
	Close() error
	BeginTransaction() (*repository.Transaction, error)

	UserExists(tx *repository.Transaction, id int64) (bool, error)
	GetUserBalance(tx *repository.Transaction, id int64, forUpdate bool) (repository.Balance, error)
	GetUserHistory(tx *repository.Transaction, id int64) ([]repository.Transfer, error)
	CreateUser(tx *repository.Transaction, id int64) error
	ChangeUserBalance(tx *repository.Transaction, id int64, amount int64) error
	UpdateHistory(tx *repository.Transaction, id int64, amount int64, purpose string) error
}

type BalanceService struct {
	repo Repository
}

func New(repository Repository) *BalanceService {
	return &BalanceService{repo: repository}
}

//Helper function for accessing database. Since other functions such as update balance need to
//access database for actual values but do not want to start new transaction, there is this function
func (bs *BalanceService) getBalance(tx *repository.Transaction, id int64, currency string, forUpdate bool) (*Balance, error) {
	secondaryBalance, err := bs.repo.GetUserBalance(tx, id, forUpdate)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrUserNotFound
		}

		return nil, ErrAccessDatabase
	}

	//convert value stored in database to needed currency
	if currency != defaultCurrency {
		url := fmt.Sprintf("https://free.currconv.com/api/v7/convert?q=%s_%s&compact=ultra&apiKey=%s", defaultCurrency, currency, exchangeApiKey)
		response, err := http.Get(url)
		if err != nil || response.StatusCode != http.StatusOK {
			return nil, ErrConvertCurrency
		}

		jsonBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, ErrConvertCurrency
		}

		exchangeRateValue := gjson.Get(string(jsonBytes), fmt.Sprintf("%s_%s", defaultCurrency, currency))
		exchangeRate := exchangeRateValue.Float()

		if exchangeRate > 1.0 && int64(float64(math.MaxInt64)/exchangeRate) < secondaryBalance.Amount {
			return nil, ErrBalanceOverflow
		}

		secondaryBalance = repository.Balance{Amount: int64(float64(secondaryBalance.Amount) * exchangeRate)}
	}

	balance, err := newBalance(secondaryBalance.Amount, currency)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

//params must be validated:
//id mist be >= 0
func (bs *BalanceService) GetBalance(id int64, currency string) (*Balance, error) {
	tx, err := bs.repo.BeginTransaction()
	if err != nil {
		return nil, ErrAccessDatabase
	}
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	balance, err := bs.getBalance(tx, id, currency, false)
	return balance, err
}

//params must be validated:
//id mist be >= 0
func (bs *BalanceService) GetHistory(id int64) ([]*Transfer, error) {
	tx, err := bs.repo.BeginTransaction()
	if err != nil {
		return nil, ErrAccessDatabase
	}
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	exists, err := bs.repo.UserExists(tx, id)
	if err != nil {
		return nil, ErrAccessDatabase
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	transfers := make([]*Transfer, 0)
	dbTransfers, err := bs.repo.GetUserHistory(tx, id)
	if err != nil {
		return nil, ErrAccessDatabase
	}

	for _, dbTransfer := range dbTransfers {
		transfer, err := newTransfer(dbTransfer.Amount, dbTransfer.TransferredAt, dbTransfer.Purpose)
		if err != nil {
			return nil, err
		}

		transfers = append(transfers, transfer)
	}
	return transfers, nil
}

//params must be validated:
//id must be >= 0
func (bs *BalanceService) ChangeBalance(id int64, amount int64) (*Balance, error) {
	tx, err := bs.repo.BeginTransaction()
	if err != nil {
		return nil, ErrAccessDatabase
	}
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	balanceStruct, err := bs.getBalance(tx, id, defaultCurrency, true)

	if err != nil {
		if err != ErrUserNotFound {
			return nil, err
		}

		//user doesn't exist, try to create new account
		if amount < 0 {
			return nil, ErrCreatingWithNegativeAmount
		}

		//new account is created with first money crediting
		err = bs.repo.CreateUser(tx, id)
		if err != nil {
			return nil, ErrAccessDatabase
		}

		balanceStruct, err = bs.getBalance(tx, id, defaultCurrency, true)
		if err != nil {
			return nil, err
		}
		//user was successfully created, continue money processing
	}

	//user is already existing here

	balanceSecondary, err := balanceStruct.ConvertToSecondary()
	if err != nil {
		return nil, err
	}

	if amount < 0 && balanceSecondary+amount < 0 {
		return nil, ErrNotEnoughMoney
	}

	if amount > 0 && math.MaxInt64-amount < balanceSecondary {
		return nil, ErrBalanceOverflow
	}

	if amount != 0 {
		err = bs.repo.ChangeUserBalance(tx, id, amount)
		if err != nil {
			return nil, ErrAccessDatabase
		}

		err = bs.repo.UpdateHistory(tx, id, amount, "External service operation")
		if err != nil {
			return nil, ErrAccessDatabase
		}
	}

	//get record and return successfully
	balanceStruct, err = bs.getBalance(tx, id, defaultCurrency, false)
	if err != nil {
		return nil, err
	}

	return balanceStruct, nil
}

//params must be validated:
//both ids should be >= 0, ids should not be equal, amount should be positive value
func (bs *BalanceService) Transfer(senderId int64, recipientId int64, amount int64) (*Balance, error) {
	tx, err := bs.repo.BeginTransaction()
	if err != nil {
		return nil, ErrAccessDatabase
	}
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	senderBalanceStruct, err := bs.getBalance(tx, senderId, defaultCurrency, true)
	if err != nil {
		return nil, err
	}
	senderBalance, err := senderBalanceStruct.ConvertToSecondary()
	if err != nil {
		return nil, err
	}
	if senderBalance < amount {
		return nil, ErrNotEnoughMoney
	}

	recipientBalanceStruct, err := bs.getBalance(tx, recipientId, defaultCurrency, true)
	if err != nil {
		//new account is not created, user cannot transfer money to
		//non-existing person
		return nil, err
	}
	recipientBalance, err := recipientBalanceStruct.ConvertToSecondary()
	if err != nil {
		return nil, err
	}
	if math.MaxInt64-amount < recipientBalance {
		return nil, ErrBalanceOverflow
	}

	err = bs.repo.ChangeUserBalance(tx, senderId, amount*-1)
	if err != nil {
		return nil, err
	}
	err = bs.repo.UpdateHistory(tx, senderId, amount*-1, fmt.Sprintf("Transferred to %d", recipientId))
	if err != nil {
		return nil, ErrAccessDatabase
	}

	err = bs.repo.ChangeUserBalance(tx, recipientId, amount)
	if err != nil {
		return nil, err
	}
	err = bs.repo.UpdateHistory(tx, recipientId, amount, fmt.Sprintf("Transferred from %d", senderId))
	if err != nil {
		return nil, ErrAccessDatabase
	}

	senderBalanceStruct, err = bs.getBalance(tx, senderId, defaultCurrency, false)
	if err != nil {
		return nil, err
	}
	return senderBalanceStruct, nil
}
