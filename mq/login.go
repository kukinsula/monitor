package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type SigninParams struct {
	Login      string `json:"login" binding:"required"`
	Password   string `json:"password" binding:"required"`
	ResponseId string
	TS         int64
}

type SigninResult struct {
	Authenticated bool
	Token         string
	TS            int64
}

func (params *SigninParams) String() string {
	return fmt.Sprintf("{%s %s %d}", params.Login, params.Password, params.TS)
}

func (result *SigninResult) String() string {
	return fmt.Sprintf("{%t %s %d}", result.Authenticated, result.Token, result.TS)
}

func (service *Service) Signin(
	ctx context.Context,
	params SigninParams) (*SigninResult, error) {

	params.TS = time.Now().UnixNano()

	fmt.Printf("REQ %s -> %s\n", LOGIN_SIGNIN, params.String())

	data, err := service.codec.Marshal(params)
	if err != nil {
		return nil, err
	}

	reply := make(chan *SigninResult, 1)
	defer close(reply)

	err = service.request(ctx, LOGIN_SIGNIN, data,

		func(message *redis.Message) error {
			result := &SigninResult{TS: time.Now().UnixNano()}

			err := service.codec.Unmarshal(message.Data, result)
			if err != nil {
				return err
			}

			fmt.Printf("RESP %s <- %s\n", LOGIN_SIGNIN, result)

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

			fmt.Printf("REQ %s <- %s\n", LOGIN_SIGNIN, params)

			result := fn(params)

			fmt.Printf("RESP %s -> %s\n", LOGIN_SIGNIN, result)

			return service.codec.Marshal(result)
		})
}
