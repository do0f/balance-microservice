package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	driverName     = "pgx"
	dataSourceName = "postgres://postgres:secret@localhost:8080/postgres"
)

var db *sql.DB

func Open() error {
	var err error
	db, err = sql.Open(driverName, dataSourceName)
	return err
}

func Close() error {
	err := db.Close()
	return err
}

func BeginTransaction() error {
	_, err := db.Exec("BEGIN")
	return err
}
func CommitTransaction() error {
	_, err := db.Exec("COMMIT")
	return err
}

func GetUserBalance(id int64, forUpdate bool) (int64, error) {
	var accessType string
	if forUpdate {
		accessType = "FOR UPDATE"
	} else {
		accessType = "FOR SHARE"
	}

	query := fmt.Sprintf("SELECT balance FROM UserBalance WHERE id = %d %s", id, accessType)
	row := db.QueryRow(query)

	var balance int64 = 0
	err := row.Scan(&balance)
	return balance, err
}

func CreateUser(id int64) error {
	_, err := db.Exec(fmt.Sprintf("INSERT INTO UserBalance (id, balance) VALUES (%d, 0)", id))
	return err
}

func ChangeUserBalance(id int64, amount int64) error {
	_, err := db.Exec(fmt.Sprintf("UPDATE UserBalance SET balance = balance + %d WHERE id = %d", amount, id))
	return err
}
