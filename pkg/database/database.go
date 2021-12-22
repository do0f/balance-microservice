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

func GetBalance(id int64) (int64, error) {
	row := db.QueryRow(fmt.Sprintf("SELECT balance FROM UserBalance WHERE id = %d", id))

	var balance int64
	err := row.Scan(&balance)
	if err != nil {
		return 0, err
	}

	return balance, nil
}
