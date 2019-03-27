package main

import (
	"context"
	"fmt"

	"github.com/kukinsula/monitor/mq"
)

var service *mq.Service

func main() {
	fmt.Print("LOGIN: connecting...")

	service = mq.NewService(":6379")
	defer service.Close()

	fmt.Println("\nLOGIN: successfully connected!")

	_, err := service.Ping()
	if err != nil {
		fmt.Printf("LOGIN: Error cannot 'PING': %v", err)
	}

	ctx := context.TODO()
	err = service.HandleSignin(ctx)
	if err != nil {
		fmt.Printf("LOGIN: Error cannot 'HandleSignin': %v", err)
	}
}
