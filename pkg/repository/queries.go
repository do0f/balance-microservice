package repository

import (
	"fmt"
	"time"
)

func (postgres *Postgres) UserExists(tx *Transaction, id int64) (bool, error) {
	row := postgres.db.QueryRow("SELECT EXISTS(SELECT balance from UserBalance where id = $1)", id)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

func (postgres *Postgres) GetUserBalance(tx *Transaction, id int64, forUpdate bool) (Balance, error) {
	var accessType string
	if forUpdate {
		accessType = "FOR UPDATE"
	} else {
		accessType = "FOR SHARE"
	}

	query := fmt.Sprintf("SELECT balance FROM UserBalance WHERE id = %d %s", id, accessType)
	row := tx.QueryRow(query)

	var balance int64 = 0
	err := row.Scan(&balance)
	return Balance{balance}, err
}

func (postgres *Postgres) GetUserHistory(tx *Transaction, id int64) ([]Transfer, error) {
	rows, err := tx.Query("SELECT amount, transferred_at, purpose FROM UserTransfers where id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transfers := make([]Transfer, 0)

	for rows.Next() {
		var transfer Transfer
		err = rows.Scan(&transfer.Amount, &transfer.TransferredAt, &transfer.Purpose)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	return transfers, nil
}

func (postgres *Postgres) CreateUser(tx *Transaction, id int64) error {
	_, err := tx.Exec("INSERT INTO UserBalance (id, balance) VALUES ($1, 0)", id)
	return err
}

func (postgres *Postgres) ChangeUserBalance(tx *Transaction, id int64, amount int64) error {
	_, err := tx.Exec("UPDATE UserBalance SET balance = balance + $1 WHERE id = $2", amount, id)
	return err
}

func (postgres *Postgres) UpdateHistory(tx *Transaction, id int64, amount int64, purpose string) error {
	_, err := tx.Exec("INSERT INTO UserTransfers (id, amount, transferred_at, purpose) VALUES ($1, $2, $3, $4)",
		id, amount, time.Now(), purpose)
	return err
}
