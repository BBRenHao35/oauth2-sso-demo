package main

import (
	"oauth2-sso-demo/config"
	"oauth2-sso-demo/handler"
	"oauth2-sso-demo/jwks"
	"oauth2-sso-demo/middleware"
	"oauth2-sso-demo/store"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()
	store.Init()
	jwks.Init(config.Cfg.KeycloakBase + "/certs")

	r := gin.Default()
	r.SetTrustedProxies(nil)

	// CORS
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

	// Token Refresh：每個 request 都自動檢查，快過期就靜默換新
	r.Use(middleware.TokenRefresh())

	// 一般路由
	r.GET("/api/auth/login", handler.Login)
	r.GET("/api/auth/callback", handler.Callback)
	r.GET("/api/auth/logout", handler.Logout)
	r.GET("/api/auth/me", handler.Me)

	// 示範 role-based 保護的 API
	r.GET("/api/admin/data", middleware.RequireRole("Admin", "Advanced"), handler.AdminData)

	r.Run(":8081")
}
