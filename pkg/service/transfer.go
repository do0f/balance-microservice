package service

import "time"

type Transfer struct {
	PrimaryValue   int64     `json:"primary value"`
	SecondaryValue int8      `json:"secondary value"`
	TransferredAt  time.Time `json:"transferred at"`
	Purpose        string    `json:"purpose"`
}

func newTransfer(secondary int64, transferredAt time.Time, purpose string) (*Transfer, error) {
	return &Transfer{secondary / 100, int8(secondary % 100), transferredAt, purpose}, nil
}
