package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	"github.com/huimingz/hyweb-assessment/config"
	"github.com/huimingz/hyweb-assessment/handler"
	"github.com/huimingz/hyweb-assessment/middleware"
)

func New(h *handler.Handler, cfg *config.Config, logger *slog.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(logger))

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
		}

		user := v1.Group("/user")
		user.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			user.PUT("/password", h.ChangePassword)
		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
