package main

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

func (api *API) Streaming(ctx *gin.Context) {
	streamer := NewStreamer(ctx)

	api.addStreamer <- streamer

	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")

	ctx.Writer.WriteHeader(200)

	ctx.Stream(func(writer io.Writer) bool {
		if data, ok := <-streamer.channel; ok {
			ctx.SSEvent("metrics", data)
			return true
		}

		return false
	})

	api.removeStreamer <- streamer
}
