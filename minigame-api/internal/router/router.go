package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"fpxxl/minigame-api/internal/config"
	"fpxxl/minigame-api/internal/handler"
	"fpxxl/minigame-api/internal/middleware"
	"fpxxl/minigame-api/internal/repository"
)

func New(cfg config.Config, repo repository.Repository, h *handler.Handler) http.Handler {
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), cors(cfg.AllowOrigin))

	r.GET("/healthz", h.Health)

	v1 := r.Group("/api/v1")
	v1.POST("/auth/login", h.Login)
	v1.GET("/config/init", h.InitConfig)
	v1.GET("/config/levels", h.Levels)
	v1.GET("/config/ads", h.AdConfig)
	v1.GET("/leaderboard", h.Leaderboard)

	authed := v1.Group("")
	authed.Use(middleware.Auth(repo))
	authed.GET("/player/progress", h.Progress)
	authed.POST("/levels/start", h.StartLevel)
	authed.POST("/levels/results", h.SubmitLevelResult)
	authed.POST("/tools/change", h.ChangeToolCount)
	authed.GET("/shop/products", h.ShopProducts)
	authed.POST("/shop/purchase", h.Purchase)
	authed.POST("/events/batch", h.Events)

	return r
}

func cors(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Openid")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
