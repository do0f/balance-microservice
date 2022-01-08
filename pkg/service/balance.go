package service

import "math"

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
