package mq

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/satori/go.uuid"
)

const responseIdLen = 16 // bytes

type MessageCallback func(message *redis.Message) error
type RequestCallback func(id string, message *redis.Message) error

type client struct {
	pool *redis.Pool
}

type SubscribeParams struct {
	Context      context.Context
	Channel      Channel
	OnSubscribed func() error
	OnMessage    MessageCallback
	Ping         time.Duration
}

func newClient(address string) *client {
	return &client{
		pool: &redis.Pool{
			MaxActive:   10,
			MaxIdle:     5,
			IdleTimeout: 200 * time.Second,
			Wait:        true,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", address)
			},
		},
	}
}

func (client *client) Close() error {
	return client.pool.Close()
}

func (client *client) Ping() (string, error) {
	conn := client.pool.Get()
	defer conn.Close()

	return redis.String(conn.Do("PING"))
}

func (client *client) publish(channel Channel, data []byte) error {
	conn := client.pool.Get()
	defer conn.Close()

	_, err := conn.Do("PUBLISH", string(channel), data)

	return err
}

func (client *client) subscribe(params SubscribeParams) error {
	conn := client.pool.Get()
	defer conn.Close()

	pubsub := redis.PubSubConn{Conn: conn}

	err := pubsub.Subscribe(string(params.Channel))
	if err != nil {
		return err
	}

	done := make(chan error)

	go client.receive(pubsub, done, params.OnSubscribed, params.OnMessage)

	ticker := time.NewTicker(params.Ping)
	defer ticker.Stop()

	for goOn := true; goOn; goOn = goOn && err == nil {
		select {
		case err = <-done: // Receive goroutine failed or unsubscription ended

		case <-ticker.C: // Connection health check
			err = pubsub.Ping("")

		case <-params.Context.Done(): // Canceled by caller
			goOn = false
		}
	}

	pubsub.Unsubscribe(string(params.Channel))

	if err == nil {
		err = <-done
	}

	return err
}

func (client *client) receive(
	pubsub redis.PubSubConn,
	done chan error,
	onSubscribed func() error,
	onMessage MessageCallback) {

	var err error

	for goOn := true; goOn; goOn = goOn && err == nil {
		switch resp := pubsub.Receive().(type) {
		case error:
			err = resp

		case redis.Subscription:
			switch resp.Count {
			case 0:
				goOn = false

			case 1:
				if onSubscribed != nil {
					err = onSubscribed()
				}
			}

		case redis.Message:
			if onMessage != nil {
				err = onMessage(&resp)
			}
		}
	}

	done <- err
	close(done)
}

func (client *client) request(
	ctx context.Context,
	channel Channel,
	data []byte,
	fn MessageCallback) error {

	responseId := uuid.Must(uuid.NewV4()).Bytes()
	ctx2, cancel := context.WithCancel(ctx)
	response := make(chan *redis.Message)
	done := make(chan error)

	go func() {
		done <- client.subscribe(SubscribeParams{
			Context: ctx2,
			Channel: Channel(responseId),

			OnSubscribed: func() error {
				conn := client.pool.Get()
				defer conn.Close()

				_, err := conn.Do("RPUSH", string(channel), append(responseId, data...))

				return err
			},

			OnMessage: func(message *redis.Message) error {
				response <- message

				cancel() // Stops the response subscription

				return nil
			},
			Ping: time.Minute,
		})

		close(done)
		close(response)
	}()

	var err error

	select {
	case message := <-response:
		err = fn(message)

	case err = <-done: // subscribe failed

	case <-ctx.Done(): // request is Done (maybe a cancel)
	}

	return err
}

func (client *client) respond(
	ctx context.Context,
	channel Channel,
	fn func(data []byte) ([]byte, error)) error {

	conn := client.pool.Get()
	defer conn.Close()

	for {
		values, err := redis.Values(conn.Do("BLPOP", string(channel), 0))
		if err != nil {
			return err
		}

		var tmp string
		var raw []byte

		_, err = redis.Scan(values, &tmp, &raw)
		if err != nil {
			return err
		}

		responseId := string(raw[:responseIdLen])
		params := raw[responseIdLen:]

		result, err := fn(params)
		if err != nil {
			return err
		}

		err = client.publish(Channel(string(responseId)), result)

		if err != nil {
			return err
		}
	}

	return nil
}
