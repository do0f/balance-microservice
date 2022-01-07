package service

import (
	"errors"
	"math"
	"time"
)

var (
	ErrAccessDatabase             = errors.New("error while accessing database")
	ErrUserNotFound               = errors.New("user with such id doesn't exist")
	ErrNegativeBalance            = errors.New("negative balance value")
	ErrNotEnoughMoney             = errors.New("trying to withdraw more money than account has")
	ErrCreatingWithNegativeAmount = errors.New("trying to withdraw money from non-existing account")
	ErrBalanceOverflow            = errors.New("balance overflow")
	ErrConvertCurrency            = errors.New("error converting to currency")
)

const (
	defaultCurrency = "RUB"
	exchangeApiKey  = "9cfd3862d7643187e74d"
)

type Balance struct {
	PrimaryValue   int64  `json:"primary value"`
	SecondaryValue int8   `json:"secondary value"`
	Currency       string `json:"currency"`
}

type Transfer struct {
	PrimaryValue   int64     `json:"primary value"`
	SecondaryValue int8      `json:"secondary value"`
	TransferredAt  time.Time `json:"transferred at"`
	Purpose        string    `json:"purpose"`
}

func newBalance(secondary int64, currency string) (*Balance, error) {
	if secondary < 0 {
		return nil, ErrNegativeBalance
	}

	return &Balance{secondary / 100, int8(secondary % 100), currency}, nil
}

func newTransfer(secondary int64, transferredAt time.Time, purpose string) (*Transfer, error) {
	return &Transfer{secondary / 100, int8(secondary % 100), transferredAt, purpose}, nil
}

func (balance *Balance) ConvertToSecondary() (int64, error) {
	//signed overflow
	if balance.PrimaryValue > math.MaxInt64/100 {
		return 0, ErrBalanceOverflow
	}

	return balance.PrimaryValue*100 + int64(balance.SecondaryValue), nil
}