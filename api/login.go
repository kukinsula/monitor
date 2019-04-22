package main

import (
	"github.com/kukinsula/monitor/mq"

	"github.com/gin-gonic/gin"
)

func (api *API) Signin(ctx *gin.Context) {
	var params mq.SigninParams

	err := ctx.ShouldBindJSON(&params)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "BAD_JSON_BODY"})
		return
	}

	result, err := api.service.Signin(ctx, params)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "SIGNIN_UNAVAILABLE"})
		return
	}

	ctx.JSON(200, gin.H{
		"access-token": result.Token,
	})
}
