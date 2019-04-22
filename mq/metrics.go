package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Metrics struct {
	CPU map[string]interface{} `json:"cpu"`
	RAM map[string]interface{} `json:"ram"`
}

func (metrics *Metrics) String() string {
	data, err := json.Marshal(metrics)
	if err != nil {
		return ""
	}

	return string(data)
}

func (service *Service) PublishMetrics(metrics *Metrics) error {
	data, err := service.codec.Marshal(metrics)
	if err != nil {
		return err
	}

	fmt.Printf("PUB %s -> %s\n", METRICS, metrics)

	return service.publish(METRICS, data)
}

func (service *Service) SubscribeToMetrics(
	ctx context.Context,
	reply chan *Metrics) error {

	err := service.subscribe(SubscribeParams{
		Context: ctx,
		Channel: METRICS,
		Ping:    time.Minute,

		OnMessage: func(message *redis.Message) error {
			metrics := &Metrics{}

			err := service.codec.Unmarshal(message.Data, metrics)
			if err != nil {
				return err
			}

			fmt.Printf("SUB %s <- %s\n", METRICS, metrics)

			reply <- metrics

			return nil
		},
	})

	return err
}
