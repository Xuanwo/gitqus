package api

import (
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func init() {
	router = gin.Default()

	router.POST("/v0/comments/:provider/:user/:repo/:branch", CreateComment)
}
