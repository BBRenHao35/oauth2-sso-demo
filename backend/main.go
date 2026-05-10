package main

import (
	"oauth2-sso-demo/config"
	"oauth2-sso-demo/handler"
	"oauth2-sso-demo/store"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()
	store.Init()

	r := gin.Default()

	// 允許前端 localhost:5173 跨域存取
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/api/auth/login", handler.Login)
	r.GET("/api/auth/callback", handler.Callback)
	r.GET("/api/auth/logout", handler.Logout)
	r.GET("/api/auth/me", handler.Me)

	r.Run(":8081")
}
