package main

import (
	"balance/pkg/repository"
	"balance/pkg/server"
	"balance/pkg/service"
	"fmt"
	"log"
)

func main() {
	postgres := repository.New()
	err := postgres.Open()
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return
	}
	defer postgres.Close()

	service := service.New(postgres)

	server := server.New(service)
	if err := server.Start(1323); err != nil {
		log.Fatal("failed to start server")
	}
}
