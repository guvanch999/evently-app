package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/guvanch999/evently-app/cmd"
	"github.com/guvanch999/evently-app/docs"
	"github.com/guvanch999/evently-app/internal"
	"github.com/guvanch999/evently-app/internal/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

//	@title						Project base API
//	@version					1.0
//	@contact.name				API Support
//	@contact.email				thelaunchist@gmail.com
//	@BasePath					/v1
//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				Bearer token is required to access the SendGoblin API. If you don't have a token, please register in the system first

// @Tag.name Auth
// @Tag.description Auth operations

func main() {
	server := cmd.NewServer()

	if server.Config.Environment == "development" {
		docs.SwaggerInfo.Schemes = []string{"http"}
		docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%v", server.Config.Port)
	} else {
		docs.SwaggerInfo.Schemes = []string{"https"}
		docs.SwaggerInfo.Host = strings.ReplaceAll(strings.ReplaceAll(server.Config.AppUrl, "https://", ""), "/v1", "")
	}

	server.Echo.GET("/swagger/*", echoSwagger.EchoWrapHandler())
	middleware.RegisterMiddlewares(server)
	internal.InitRouters(server)

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go func() {
		if err := server.Echo.Start(fmt.Sprintf(":%v", server.Config.Port)); err != nil && !errors.Is(
			err,
			http.ErrServerClosed,
		) {
			server.Echo.Logger.Fatal("shutting down the server. error: " + err.Error())
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	<-signalCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Echo.Shutdown(ctx); err != nil {
		server.Echo.Logger.Fatal(err)
	}
}
