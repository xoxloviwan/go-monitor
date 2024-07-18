package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//fmt.Printf("%v+\n", ctx)
		fmt.Println("The URL: ", ctx.Request.URL.Path)
		ctx.Next()
	}
	//return gin.LoggerWithWriter(gin.DefaultWriter, "[api] ")
}
