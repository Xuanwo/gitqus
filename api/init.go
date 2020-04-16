package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

func init() {
	router = gin.Default()

	router.POST("/v0/comments/:provider/:user/:repo/:branch", CreateComment)
}
