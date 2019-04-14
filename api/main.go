package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"

	"context"
	"time"

	"github.com/kukinsula/monitor/mq"
)

func main() {
	fmt.Println("Connecting...")

	service, err := mq.NewService(":6379")
	if err != nil {
		fmt.Printf("Redis connection failed %v", err)
		return
	}

	fmt.Println("Successfully connected!")

	go func() {
		for {
			params := mq.SigninParams{
				Login:    "Albert",
				Password: "Binc",
				TS:       time.Now().UnixNano(),
			}

			_, err := service.Signin(context.TODO(), params)
			if err != nil {
				fmt.Printf("API error cannot request: %+v\n", err)
				return
			}

			time.Sleep(time.Duration(random(100, 500)) * time.Millisecond)
		}
	}()

	quit := make(chan os.Signal, 1)
	cleanup := make(chan struct{}, 1)

	signal.Notify(quit, os.Interrupt)

	defer close(quit)
	defer close(cleanup)

	api := NewAPI(service)

	go api.Run()

	go func() {
		<-quit

		fmt.Println("Quiting...")

		err := api.Shutdown()
		if err != nil {
			fmt.Println("API Shutdown failed:", err)
		}

		service.Close()
		if err != nil {
			fmt.Println("API Shutdown failed:", err)
		}

		cleanup <- struct{}{}
	}()

	<-cleanup

	fmt.Println("Done")
}

func random(min, max int) int64 {
	return rand.Int63n(int64(max-min)) + int64(min)
}
