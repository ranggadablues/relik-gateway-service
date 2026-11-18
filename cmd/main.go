package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"relik-gateway-service/internal/transport"
	"relik-gateway-service/middleware"

	appEmbed "relik-gateway-service"

	"github.com/joho/godotenv"
	logger "github.com/ranggadablues/gosok/logger"
)

func main() {
	log := logger.NewLogger()

	// Load .env into environment
	//
	dir, err := os.Getwd()
	if err != nil {
		log.Error("msg", "Dir not found with error: "+err.Error())
	}
	env := fmt.Sprintf("%s/.env", dir)
	if err := godotenv.Load(env); err != nil {
		log.Error("msg", "No .env file found: "+err.Error())
	}

	if err := middleware.InitAuth(); err != nil {
		log.Error("msg", "failed to init auth: "+err.Error())
		panic(err)
	}

	handler := transport.NewHTTPHandler(
		transport.SwaggerAssets{
			JSON: appEmbed.SwaggerJSON,
			UI:   appEmbed.SwaggerUI,
		},
	)

	gatewayPort := os.Getenv("GATEWAY_SERVICE_PORT")
	gatewayRunPort := fmt.Sprintf(":%s", gatewayPort)
	log.Info("Gateway service running", "addr", gatewayRunPort)

	srv := &http.Server{
		Addr:         gatewayRunPort,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("msg", err.Error())
		panic(err)
	}
}
