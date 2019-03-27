package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Service struct {
	*client
	codec Codec
}

func NewService(address string) *Service {
	return &Service{
		client: newClient(address),
		codec:  &JsonCodec{},
	}
}

func (service *Service) PublishMetrics(metrics *Metrics) error {
	data, err := service.codec.Marshal(metrics)
	if err != nil {
		return err
	}

	return service.publish(METRICS, data)
}

func (service *Service) SubscribeToMetrics(
	ctx context.Context,
	reply chan *Metrics) error {

	err := service.subscribe(ctx, METRICS,

		func(message *redis.Message) error {
			metrics := &Metrics{}

			err := service.codec.Unmarshal(message.Data, metrics)
			if err != nil {
				return err
			}

			reply <- metrics

			return nil
		},
		time.Minute)

	if err != nil {
		return err
	}

	return nil
}

type SigninParams struct {
	Login      string
	Password   string
	ResponseId string
	TS         int64
}

type SigninResult struct {
	Authenticated bool
	Token         string
}

func (service *Service) Signin(
	ctx context.Context,
	params SigninParams) (*SigninResult, error) {

	data, err := service.codec.Marshal(params)
	if err != nil {
		return nil, err
	}

	reply := make(chan *SigninResult, 1)
	defer close(reply)

	err = service.request(ctx, LOGIN_SIGNIN, data,

		func(message *redis.Message) error {
			result := &SigninResult{}

			err := service.codec.Unmarshal(message.Data, result)
			if err != nil {
				return err
			}

			reply <- result

			return nil
		})

	if err != nil {
		return nil, err
	}

	return <-reply, nil
}

func (service *Service) HandleSignin(
	ctx context.Context) error {

	return service.respond(ctx, LOGIN_SIGNIN,

		func(data []byte) ([]byte, error) {
			params := &SigninParams{}

			err := service.codec.Unmarshal(data, params)
			if err != nil {
				return nil, err
			}

			fmt.Println("HANDLE SIGNIN", params)

			// TODO: recevoir un callback qui recevra params et traitera la
			// requêtes (chrcher en base le user etc..).

			result := SigninResult{
				Authenticated: true,
				Token:         "AZERTY.1234.QWERTY",
			}

			return service.codec.Marshal(result)
		})
}
