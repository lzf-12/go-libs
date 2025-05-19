package rest

import (
	"github.com/gin-gonic/gin"
)

func InitGin() *gin.Engine {
	return gin.New()
}

func HandlerGin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gin.ResponseWriter.WriteString(ctx.Writer, "Hello from gin rest!")
	}
}
