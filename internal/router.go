package internal

import (
	"github.com/guvanch999/evently-app/cmd"
	"github.com/guvanch999/evently-app/internal/handlers"
	"github.com/guvanch999/evently-app/internal/repositories"
	"github.com/guvanch999/evently-app/internal/services"
)

func InitRouters(server *cmd.Server) {
	userRepository := repositories.NewUserRepository(server.Database)
	authService := services.NewAuthService(server.Config, userRepository)
	authHandler := handlers.NewAuthHandler(authService)

	authRouter := server.Echo.Group("/auth")

	authRouter.GET("/google", authHandler.HandleGoogleLogin)
	authRouter.GET("/callback", authHandler.HandleGoogleCallback)
}
