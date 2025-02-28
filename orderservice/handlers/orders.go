package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"orderservice/models"
	"orderservice/repository"
)

type OrderHandler struct {
	repo   *repository.OrderRepository
	logger *zap.Logger
}

func NewOrderHandler(db *gorm.DB, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		repo:   repository.NewOrderRepository(db),
		logger: logger,
	}
}

func (h *OrderHandler) CreateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		h.logger.Error("Invalid customer input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CreateCustomer(&customer); err != nil {
		h.logger.Error("Failed to create customer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Customer creation failed"})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		h.logger.Error("Invalid order input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CreateOrder(&order); err != nil {
		h.logger.Error("Order creation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order creation failed"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var status struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&status); err != nil {
		h.logger.Error("Invalid status input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateOrderStatus(uint(id), status.Status); err != nil {
		h.logger.Error("Status update failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Status update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Error("Invalid order ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := h.repo.GetOrder(uint(id))  // Changed from GetOrderByID to GetOrder
	if err != nil {
		h.logger.Error("Failed to fetch order", zap.Error(err), zap.Int("id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		h.logger.Error("Invalid product input", zap.Error(err))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CreateProduct(&product); err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		c.JSON(500, gin.H{"error": "Product creation failed"})
		return
	}

	c.JSON(201, product)
}

func (h *OrderHandler) GetProducts(c *gin.Context) {
	var products []models.Product
	if err := h.repo.GetAllProducts(&products); err != nil {
		h.logger.Error("Failed to fetch products", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to retrieve products"})
		return
	}

	c.JSON(200, products)
}