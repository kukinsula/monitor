package mq

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type SigninParams struct {
	Login      string
	Password   string
	ResponseId string
	TS         int64
}

type SigninResult struct {
	Authenticated bool
	Token         string
	TS            int64
}

func (s *SigninParams) String() string {
	return fmt.Sprintf("{%s %s %d}", s.Login, s.Password, s.TS)
}

func (s *SigninResult) String() string {
	return fmt.Sprintf("{%t %s %d}", s.Authenticated, s.Token, s.TS)
}

func (service *Service) Signin(
	ctx context.Context,
	params SigninParams) (*SigninResult, error) {

	fmt.Printf("-> %v\n", params)

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

			fmt.Printf("<- %v\n", result)

			reply <- result

			return nil
		})

	if err != nil {
		return nil, err
	}

	return <-reply, nil
}

func (service *Service) HandleSignin(
	ctx context.Context,
	fn func(params *SigninParams) *SigninResult) error {

	return service.respond(ctx, LOGIN_SIGNIN,

		func(data []byte) ([]byte, error) {
			params := &SigninParams{}

			err := service.codec.Unmarshal(data, params)
			if err != nil {
				return nil, err
			}

			fmt.Printf("<- %v\n", params)

			result := fn(params)

			fmt.Printf("-> %v\n", result)

			return service.codec.Marshal(result)
		})
}
