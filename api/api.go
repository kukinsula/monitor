package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

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
	defer close(quit)

	signal.Notify(quit, os.Interrupt)
	cleanup := make(chan struct{}, 1)

	go func() {
		<-quit

		fmt.Println("Quiting...")

		cancel()
		close(reply)

		cleanup <- struct{}{}
		close(cleanup)
	}()

	go func() {
		for range reply {
		}
	}()

	go func() {
		err := service.SubscribeToMetrics(ctx, reply)
		if err != nil {
			fmt.Printf("API error cannot SubscribeToMetrics: %v", err)
			return
		}
	}()

	go func() {
		for {
			params := mq.SigninParams{
				Login:    "Albert",
				Password: "Binc",
				TS:       time.Now().UnixNano(),
			}

			_, err := service.Signin(ctx, params)
			if err != nil {
				fmt.Printf("API error cannot request: %+v\n", err)
				return
			}

			time.Sleep(time.Duration(random(100, 500)) * time.Millisecond)
		}
	}()

	<-cleanup

	fmt.Println("API: done!")
}

func random(min, max int) int64 {
	return rand.Int63n(int64(max-min)) + int64(min)
}
