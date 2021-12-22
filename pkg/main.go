package main

import (
	"balance_microservice/database"
	"balance_microservice/server"

	"fmt"
)

func main() {
	err := database.Open()
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return
	}
	defer database.Close()

	server.Start()
}
