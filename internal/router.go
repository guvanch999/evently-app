package internal

import (
	"net/http"
	"strings"

	"github.com/guvanch999/evently-app/cmd"
	"github.com/guvanch999/evently-app/internal/handlers"
	"github.com/guvanch999/evently-app/internal/middleware"
	"github.com/guvanch999/evently-app/internal/repositories"
	"github.com/guvanch999/evently-app/internal/services"
	"github.com/labstack/echo/v4"
)

func InitRouters(server *cmd.Server) {
	cfg := server.Config

	userRepository := repositories.NewUserRepository(server.Database)

	tokenService := services.NewTokenService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAccessTTL)
	authService := services.NewAuthService(cfg, userRepository, tokenService)
	userService := services.NewUserService(userRepository)

	secureCookies := strings.HasPrefix(cfg.GAuthRedirectUrl, "https://")
	authHandler := handlers.NewAuthHandler(authService, secureCookies)
	userHandler := handlers.NewUserHandler(userService)

	requireAuth := middleware.AuthMiddleware(tokenService)

	server.Echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := server.Echo.Group("/v1")

	authRouter := v1.Group("/auth/google")
	authRouter.GET("/login", authHandler.HandleGoogleLogin)
	authRouter.GET("/callback", authHandler.HandleGoogleCallback)

	meRouter := v1.Group("/me", requireAuth)
	meRouter.GET("", userHandler.GetMe)
	meRouter.PATCH("", userHandler.UpdateMe)
}
