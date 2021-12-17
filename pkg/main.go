package main

import (
	"fmt"

	echo "github.com/labstack/echo/v4"

	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	db, err := sql.Open("pgx", "postgres://postgres:secret@localhost:8080/postgres")
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return
	}
	defer db.Close()

	rows, err := db.Query("select * from UserBalance")
	if err != nil {
		fmt.Printf("Query failed: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var ub UserBalance
		err = rows.Scan(&ub.Id, &ub.Balance)
		if err != nil {
			fmt.Printf("Scan failed: %v\n", err)
			return
		}
		fmt.Println(ub.String())
	}

	e := echo.New()
	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

type UserBalance struct {
	Id      int `json:"id"`
	Balance int `json:"balance"`
}

func (ub UserBalance) String() string {
	return fmt.Sprintf("User id: %d\nBalance: %d roubles, %d kopeks\n", ub.Id, ub.Balance/100, ub.Balance%100)
}
