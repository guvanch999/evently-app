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

// Handlers groups the HTTP handlers mounted by RegisterRoutes.
type Handlers struct {
	Auth  *handlers.AuthHandler
	User  *handlers.UserHandler
	Event *handlers.EventHandler
}

func InitRouters(server *cmd.Server) {
	cfg := server.Config

	userRepository := repositories.NewUserRepository(server.Database)
	eventRepository := repositories.NewEventRepository(server.Database)

	tokenService := services.NewTokenService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAccessTTL)
	authService := services.NewAuthService(cfg, userRepository, tokenService)
	userService := services.NewUserService(userRepository)
	eventService := services.NewEventService(eventRepository)

	secureCookies := strings.HasPrefix(cfg.GAuthRedirectUrl, "https://")

	RegisterRoutes(server.Echo, Handlers{
		Auth:  handlers.NewAuthHandler(authService, secureCookies),
		User:  handlers.NewUserHandler(userService),
		Event: handlers.NewEventHandler(eventService),
	}, middleware.AuthMiddleware(tokenService))
}

// RegisterRoutes mounts the public and JWT-protected routes on e.
func RegisterRoutes(e *echo.Echo, h Handlers, requireAuth echo.MiddlewareFunc) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := e.Group("/v1")

	authRouter := v1.Group("/auth/google")
	authRouter.GET("/login", h.Auth.HandleGoogleLogin)
	authRouter.GET("/callback", h.Auth.HandleGoogleCallback)

	meRouter := v1.Group("/me", requireAuth)
	meRouter.GET("", h.User.GetMe)
	meRouter.PATCH("", h.User.UpdateMe)

	eventRouter := v1.Group("/events")
	eventRouter.GET("", h.Event.ListEvents)
	eventRouter.GET("/:id", h.Event.GetEvent)
	eventRouter.POST("", h.Event.CreateEvent, requireAuth)
	eventRouter.PATCH("/:id", h.Event.UpdateEvent, requireAuth)
	eventRouter.DELETE("/:id", h.Event.DeleteEvent, requireAuth)
}
