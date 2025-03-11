package internal

import (
	"github.com/guvanch999/evently-app/cmd"
	"github.com/guvanch999/evently-app/internal/handlers"
	"github.com/guvanch999/evently-app/internal/services"
)

func InitRouters(server *cmd.Server) {
	authService := services.NewAuthService(server.Config)
	authHandler := handlers.NewAuthHandler(authService)

	authRouter := server.Echo.Group("/auth")

	authRouter.GET("/google", authHandler.HandleGoogleLogin)
	authRouter.GET("/callback", authHandler.HandleGoogleCallback)
}
