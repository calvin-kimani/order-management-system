package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"paymentservice/handlers"
)

func main() {
	// Initialize structured logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	// Initialize payment handler with M-Pesa client
	paymentHandler := handlers.NewPaymentHandler(logger)

	// Create Gin router with middleware
	router := gin.Default()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Register payment endpoints
	router.POST("/payments", paymentHandler.ProcessPayment)
	router.POST("/callback", paymentHandler.PaymentCallback)

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default port
	}

	logger.Info("Starting Payment Service",
		zap.String("port", port),
		zap.String("environment", os.Getenv("GIN_MODE")),
	)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Fatal("Failed to start server",
			zap.Error(err),
		)
	}
}