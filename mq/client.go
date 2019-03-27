package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/satori/go.uuid"
)

const responseIdLen = 16

type MessageCallback func(message *redis.Message) error
type RequestCallback func(id string, message *redis.Message) error

type client struct {
	pool *redis.Pool
}

func newClient(address string) *client {
	return &client{
		pool: &redis.Pool{
			// MaxActive: 10,
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
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

	return redis.String(conn.Do("PING"))
}

func (client *client) Uuid() string {
	return uuid.Must(uuid.NewV4()).String()
}

func (client *client) publish(channel Channel, data []byte) error {
	conn := client.pool.Get()
	_, err := conn.Do("PUBLISH", string(channel), data)

	return err
}

func (client *client) subscribe(ctx context.Context,
	channel Channel,
	fn MessageCallback,
	ping time.Duration) error {

	conn := client.pool.Get()
	pubsub := redis.PubSubConn{Conn: conn}

	err := pubsub.Subscribe(string(channel))
	if err != nil {
		return err
	}

	done := make(chan error, 1)

	go func() {
		defer pubsub.Unsubscribe(channel)

		var err error
		for err == nil {
			switch resp := pubsub.Receive().(type) {
			case redis.Error:
				err = resp

			case redis.Message:
				err = fn(&resp)

			case redis.Subscription:
			case redis.Pong:

			default:
				err = fmt.Errorf("client.subscribe Receive returned unknow type %v", resp)
			}
		}

		done <- err
	}()

	ticker := time.NewTicker(ping)
	defer ticker.Stop()

	for goOn := true; goOn; goOn = goOn && err == nil {
		select {
		case err = <-done:

		case <-ticker.C:
			err = pubsub.Ping("")

		case <-ctx.Done():
			goOn = false
		}
	}

	close(done)

	return err
}

func (client *client) request(
	ctx context.Context,
	channel Channel,
	data []byte,
	fn MessageCallback) error {

	responseId := uuid.Must(uuid.NewV4()).Bytes()

	ctx2, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)

	go func() {
		done <- client.subscribe(ctx2, Channel(responseId),
			func(message *redis.Message) error {
				cancel() // Stops the response subscription

				fn(message)

				return nil
			},
			time.Minute)
	}()

	conn := client.pool.Get()

	_, err := conn.Do("RPUSH", string(channel), append(responseId, data...))
	if err != nil {
		return err
	}

	select {
	case err = <-done: // subscribe failed
	case <-ctx.Done(): // request is Done (maybe a cancel)
	}

	cancel()

	return err
}

func (client *client) respond(
	ctx context.Context,
	channel Channel,
	fn func(data []byte) ([]byte, error)) error {

	for {
		conn := client.pool.Get()

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
