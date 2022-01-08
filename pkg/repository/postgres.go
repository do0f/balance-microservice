package repository

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	driverName     = "pgx"
	dataSourceName = "postgres://postgres:secret@localhost:8080/postgres"
)

type Postgres struct {
	db *sql.DB
}

func New() *Postgres {
	return &Postgres{}
}

func (postgres *Postgres) Open() error {
	var err error
	postgres.db, err = sql.Open(driverName, dataSourceName)
	return err
}

func (postgres *Postgres) Close() error {
	err := postgres.db.Close()
	return err
}

func (postgres *Postgres) BeginTransaction() (*Transaction, error) {
	tx, err := postgres.db.Begin()
	if err != nil {

		return nil, err
	}
	return &Transaction{tx}, nil
}

func (Postgres *Postgres) Commit(tx *Transaction) error {
	return tx.tx.Commit()
}

func (Postgres *Postgres) Rollback(tx *Transaction) error {
	return tx.tx.Rollback()
}
