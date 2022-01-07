package repository

import (
	"database/sql"
	"time"
)

type Transaction = sql.Tx
type Transfer struct {
	Amount        int64
	TransferredAt time.Time
	Purpose       string
}
type Balance struct {
	Amount int64
}
