package handlers

import "github.com/gin-gonic/gin"

type Failure struct {
	Error string `json:"error"`
}

type Success struct {
	any
}

type StartDownloadResponse struct {
	State string `json:"state"`
}

func ResponseFailure(ctx *gin.Context, err error) {
	ctx.JSON(400, Failure{err.Error()})
}

func ResponseInternalError(ctx *gin.Context, err error) {
	ctx.JSON(500, Failure{err.Error()})
}

func ResponseSuccess(ctx *gin.Context, data any) {
	ctx.JSON(200, Success{data})
}
