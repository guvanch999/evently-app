package internal

import (
	"github.com/guvanch999/evently-app/cmd"
	"github.com/guvanch999/evently-app/internal/handlers"
)

func InitRouters(server *cmd.Server) {
	authHandler := handlers.NewAuthHandler(server.Config)

	authRouter := server.Echo.Group("/auth")

	authRouter.GET("/google", authHandler.HandleGoogleLogin)
	authRouter.GET("/callback", authHandler.HandleGoogleCallback)
}
