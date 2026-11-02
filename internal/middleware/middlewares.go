package middleware

import (
	"net/http"

	"github.com/guvanch999/evently-app/cmd"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func RegisterMiddlewares(server *cmd.Server) {
	server.Echo.Pre(echoMiddleware.RemoveTrailingSlash())
	server.Echo.Use(echoMiddleware.Recover())
	server.Echo.Use(echoMiddleware.RequestID())
	server.Echo.Use(echoMiddleware.Secure())
	server.Echo.Use(
		echoMiddleware.BodyLimitWithConfig(
			echoMiddleware.BodyLimitConfig{
				Limit: "1M",
			},
		),
	)

	server.Echo.Use(
		echoMiddleware.CORSWithConfig(
			echoMiddleware.CORSConfig{
				AllowOrigins: server.Config.CorsAllowOrigins,
				AllowMethods: []string{
					http.MethodGet,
					http.MethodPut,
					http.MethodPatch,
					http.MethodPost,
					http.MethodDelete,
					http.MethodOptions,
				},
				AllowCredentials: true,
				AllowHeaders: []string{
					echo.HeaderOrigin,
					echo.HeaderContentType,
					echo.HeaderAccept,
					echo.HeaderAuthorization,
				},
			},
		),
	)
	server.Echo.Use(
		echoMiddleware.LoggerWithConfig(
			echoMiddleware.LoggerConfig{
				Format:           "${time_custom} | ${id} | ${method} ${path} | Status: ${status} | ${latency_human} | Error: ${error}\n",
				CustomTimeFormat: "2006/01/02 15:04:05",
			},
		),
	)
	server.Echo.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(20)))
}
