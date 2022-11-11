package main

import (
	"log"
	"user_balance_microservice/internal/app/apiserver"
)

func main() {
	config := apiserver.GetConfig()

	if err := apiserver.Start(config); err != nil {
		log.Fatal(err)
	}
}
