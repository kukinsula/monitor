package main

import (
	"fmt"
	// "math/rand"
	// "time"

	"github.com/kukinsula/monitor/mq"
)

var service *mq.Service

func main() {
	fmt.Print("DATABASE: connecting...")

	service = mq.New(":6379")
	defer service.Close()

	fmt.Println("\nDATABASE: successfully connected!")

	_, err := service.Ping()
	if err != nil {
		fmt.Printf("DATABASE: Error cannot 'PING': %v", err)
	}

	channel := "METRICS"
	reply := make(chan *mq.Metrics)
	service.Subscribe(channel, reply)

	for params := range reply {
		fmt.Printf("DATABASE <- %v\n", params)
	}
}
