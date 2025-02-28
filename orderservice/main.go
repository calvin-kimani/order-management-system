package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"orderservice/handlers"
	"orderservice/models"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Database connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Database connection failed", zap.Error(err))
	}

	// Auto migrate
	db.AutoMigrate(
		&models.Customer{},
		&models.Product{},
		&models.Order{},
	)

	// Initialize handler
	orderHandler := handlers.NewOrderHandler(db, logger)

	// Router setup
	router := gin.Default()
	router.POST("/customers", orderHandler.CreateCustomer)
	router.POST("/orders", orderHandler.CreateOrder)
	router.GET("/orders/:id", orderHandler.GetOrder)
	router.PUT("/orders/:id/status", orderHandler.UpdateOrderStatus)
	router.GET("/products", orderHandler.GetProducts)
	router.POST("/products", orderHandler.CreateProduct)

	logger.Info("Starting Orders Service on :8080")
	http.ListenAndServe(":8080", router)
}