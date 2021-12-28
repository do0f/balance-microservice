package balance

import (
	"balance_microservice/database"
	"fmt"
	"math"
)

type Balance struct {
	PrimaryValue   int64  `json:"primary value"`
	SecondaryValue int8   `json:"secondary value"`
	Currency       string `json:"currency"`
}

func newBalance(secondary int64, currency string) (*Balance, error) {
	if secondary < 0 {
		return nil, ErrNegativeBalance
	}

	return &Balance{secondary / 100, int8(secondary % 100), currency}, nil
}

func (balance *Balance) ConvertToSecondary() (int64, error) {
	//signed overflow
	if balance.PrimaryValue > math.MaxInt64/100 {
		return 0, ErrBalanceOverflow
	}

	return balance.PrimaryValue*100 + int64(balance.SecondaryValue), nil
}

var (
	ErrAccessDatabase             = fmt.Errorf("error while accessing database")
	ErrUserNotFound               = fmt.Errorf("user with such id doesn't exist")
	ErrNegativeBalance            = fmt.Errorf("negative balance value")
	ErrNotEnoughMoney             = fmt.Errorf("trying to withdraw more money than account has")
	ErrCreatingWithNegativeAmount = fmt.Errorf("trying to withdraw money from non-existing account")
	ErrBalanceOverflow            = fmt.Errorf("balance overflow")
)

//Helper function for accessing database. Since other functions such as update balance need to
//access database for actual values but do not want to start new transaction, there is this function
func getBalance(tx *database.Transaction, id int64, forUpdate bool) (*Balance, error) {
	secondaryBalance, err := database.GetUserBalance(tx, id, forUpdate)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, ErrUserNotFound
		}

		return nil, ErrAccessDatabase
	}

	//RUB is default currency
	balance, err := newBalance(secondaryBalance, "RUB")
	if err != nil {
		return nil, ErrNegativeBalance
	}

	return balance, nil
}

//helper function, id mist be >= 0
func GetBalanceTransaction(id int64) (*Balance, error) {
	tx, err := database.BeginTransaction()
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

	balance, err := getBalance(tx, id, false)
	return balance, err
}

//helper function, id must be >= 0
func ChangeBalanceTransaction(id int64, amount int64) (*Balance, error) {
	tx, err := database.BeginTransaction()
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

	balanceStruct, err := getBalance(tx, id, true)

	if err != nil {
		if err != ErrUserNotFound {
			return nil, err
		}

		//user doesn't exist, try to create new account
		if amount < 0 {
			return nil, ErrCreatingWithNegativeAmount
		}

		//new account is created with first money crediting
		err = database.CreateUser(tx, id)
		if err != nil {
			return nil, ErrAccessDatabase
		}

		balanceStruct, err = getBalance(tx, id, true)
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
		err = database.ChangeUserBalance(tx, id, amount)
		if err != nil {
			return nil, ErrAccessDatabase
		}
	}

	//get record and return successfully
	balanceStruct, err = getBalance(tx, id, false)
	if err != nil {
		return nil, err
	}

	return balanceStruct, nil
}
