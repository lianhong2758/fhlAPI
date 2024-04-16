package server

import (
	"fhlApi/fhl"

	"github.com/gin-gonic/gin"
)

func RunGin(f *fhl.FHL) {
	gin.SetMode(gin.DebugMode)
	r := gin.Default() //初始化
	r.GET("/", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"msg": "hello"}) })
	r.POST("/gettopic", GetTopic(f))
	r.POST("/answer",  UpAnswer(f))
	r.Run("0.0.0.0:8080")
}
