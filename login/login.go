package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kukinsula/monitor/mq"
)

func main() {
	fmt.Println("Connecting...")

	service, err := mq.NewService(":6379")
	if err != nil {
		fmt.Printf("Redis connection failed: %v", err)
		return
	}

	defer service.Close()

	fmt.Println("Successfully connected!")

	ctx := context.TODO()

	err = service.HandleSignin(ctx,
		func(params *mq.SigninParams) *mq.SigninResult {
			return &mq.SigninResult{
				Authenticated: true,
				Token:         "AZERTY.1234.QWERTY",
				TS:            time.Now().UnixNano(),
			}
		})
	if err != nil {
		fmt.Println("HandleSignin failed: %v", err)
	}

	fmt.Println("Done")
}
