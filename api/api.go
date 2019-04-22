package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kukinsula/monitor/mq"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

type API struct {
	service        *mq.Service
	server         *http.Server
	engine         *gin.Engine
	streamers      map[string]*Streamer
	addStreamer    chan *Streamer
	removeStreamer chan *Streamer
	stopStreaming  chan struct{}
	done           chan error
}

type Streamer struct {
	Uuid    string
	context *gin.Context
	channel chan string
}

func NewStreamer(ctx *gin.Context) *Streamer {
	return &Streamer{
		Uuid:    uuid.Must(uuid.NewV4()).String(),
		context: ctx,
		channel: make(chan string),
	}
}

func (streamer *Streamer) Stream(data string) {
	streamer.channel <- data
}

func (streamer *Streamer) Close() {
	close(streamer.channel)
}

func NewAPI(service *mq.Service) *API {
	engine := gin.New()
	api := &API{
		service: service,
		server: &http.Server{
			Addr:    ":8080",
			Handler: engine,
		},
		engine:         engine,
		streamers:      make(map[string]*Streamer),
		addStreamer:    make(chan *Streamer),
		removeStreamer: make(chan *Streamer),
		stopStreaming:  make(chan struct{}, 1),
		done:           make(chan error),
	}

	go api.listenAndPropagateMetrics()

	engine.POST("/login/signin", api.Signin)
	engine.GET("/streaming", api.Streaming)

	return api
}

func (api *API) listenAndPropagateMetrics() {
	metrics := make(chan *mq.Metrics)
	failure := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())

	go func() { failure <- api.service.SubscribeToMetrics(ctx, metrics) }()

	var err error

	for goOn := true; goOn; goOn = goOn && err == nil {
		select {
		case streamer := <-api.addStreamer:
			api.streamers[streamer.Uuid] = streamer

		case streamer := <-api.removeStreamer:
			delete(api.streamers, streamer.Uuid)

		case <-api.stopStreaming:
			goOn = false

			for uuid, streamer := range api.streamers {
				streamer.Close()
				delete(api.streamers, uuid)
			}

		case metrics := <-metrics:
			data, err := json.Marshal(metrics)
			if err != nil {
				break
			}

			for _, streamer := range api.streamers {
				streamer.Stream(string(data))
			}

		case err = <-failure:
		}
	}

	cancel()

	api.done <- err
	close(api.done)
	close(failure)
	close(metrics)
}

func (api *API) Shutdown() (err error) {
	api.stopStreaming <- struct{}{}
	close(api.stopStreaming)

	failure := make(chan error)
	defer close(failure)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	go func() { failure <- api.server.Shutdown(ctx) }()

	select {
	case <-ctx.Done():
		err = fmt.Errorf("Server shutdown timeout")

	case err = <-failure:

	case err = <-api.done:
	}

	return err
}

func (api *API) Run() {
	api.engine.Run()
}
