package router

import (
	"go-openai/internal/gpt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	routes := gin.Default()

	// 配置 CORS
	// TODO: remove in production
	routes.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 設定路由
	api := routes.Group("api/")
	api.GET("/gpt", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	gptService := gpt.NewGptService()
	gptHandler := gpt.NewGptHandler(gptService)

	gpt := api.Group("gpt")
	gpt.POST("/send", gptHandler.SendContentHandler)

	gpt.POST("/stream", gptHandler.StreamHandler)

	return routes
}
