package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/kukinsula/monitor/monitor/metric"
	"github.com/kukinsula/monitor/mq"
)

func main() {
	flag.Usage = func() { usage(nil) }

	config := metric.NewConfig()

	monitoring := NewMonitoring(config)

	fmt.Println("Connecting...")

	service, err := mq.NewService(":6379")
	if err != nil {
		fmt.Printf("Redis connection failed: %v", err)
		return
	}

	defer service.Close()

	fmt.Println("Successfully connected!")

	channel := make(chan *mq.Metrics)

	go monitoring.Start(channel)

	for metrics := range channel {
		err = service.PublishMetrics(metrics)
		if err != nil {
			fmt.Printf("PUB Metrics failed: %v\n", metrics, err)
			return
		}
	}

	fmt.Println("Done")
}

func random(min, max int) int64 {
	return rand.Int63n(int64(max-min)) + int64(min)
}

func usage(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}
