package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	// "time"

	"github.com/kukinsula/monitor/mq"
)

func main() {
	fmt.Print("API: connecting...")

	service := mq.NewService(":6379")
	defer service.Close()

	fmt.Println("\nAPI: successfully connected!")

	_, err := service.Ping()
	if err != nil {
		fmt.Printf("API error cannot 'PING': %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	reply := make(chan *mq.Metrics)

	quit := make(chan os.Signal, 1)
	cleanup := make(chan struct{})
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit

		fmt.Println("Quiting...")

		cancel()
		close(reply)
		close(cleanup)
	}()

	go func() {
		for metrics := range reply {
			fmt.Printf("API <- %s\n", metrics)
		}
	}()

	go func() {
		err := service.SubscribeToMetrics(ctx, reply) //
		if err != nil {
			fmt.Printf("API error cannot SubscribeToMetrics: %v", err)
			return
		}
	}()

	// for {
	// 	params := mq.SigninParams{
	// 		Login:    "Albert",
	// 		Password: "Binc",
	// 		TS:       time.Now().UnixNano(),
	// 	}

	// 	result, err := service.Signin(ctx, params)
	// 	if err != nil {
	// 		fmt.Printf("API error cannot request: %v", err)
	// 		return
	// 	}

	// 	fmt.Println("RESULT", result)

	// 	time.Sleep(time.Duration(random(100, 300)) * time.Millisecond)
	// }

	<-cleanup

	fmt.Println("API: done!")
}

func random(min, max int) int64 {
	return rand.Int63n(int64(max-min)) + int64(min)
}
