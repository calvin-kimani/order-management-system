package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"orderservice/handlers"
	"orderservice/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Clean up any existing tables
	db.Migrator().DropTable(&models.Order{}, &models.Customer{}, &models.Product{})
	// Create fresh tables
	db.AutoMigrate(&models.Order{}, &models.Customer{}, &models.Product{})
	return db
}

func TestCreateCustomer(t *testing.T) {
	db := setupTestDB()
	logger, _ := zap.NewDevelopment()
	handler := handlers.NewOrderHandler(db, logger)

	t.Run("Create valid customer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			"POST",
			"/customers",
			strings.NewReader(`{"name":"John Doe","email":"john@example.com"}`),
		)
		c.Request.Header.Add("Content-Type", "application/json")

		handler.CreateCustomer(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.Customer
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "John Doe", response.Name)
		assert.Equal(t, "john@example.com", response.Email)
	})

	t.Run("Invalid customer data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			"POST",
			"/customers",
			strings.NewReader(`{"invalid":"json"`),
		)
		c.Request.Header.Add("Content-Type", "application/json")

		handler.CreateCustomer(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCreateOrder(t *testing.T) {
	db := setupTestDB()
	logger, _ := zap.NewDevelopment()
	handler := handlers.NewOrderHandler(db, logger)

	// Create test customer and product first
	customer := &models.Customer{Name: "Test Customer", Email: "test@example.com"}
	db.Create(customer)
	product := &models.Product{Name: "Book", Price: 29.99}
	db.Create(product)

	t.Run("Create valid order", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			"POST",
			"/orders",
			strings.NewReader(fmt.Sprintf(`{"customer_id":%d,"products":[{"id":%d}]}`, customer.ID, product.ID)),
		)
		c.Request.Header.Add("Content-Type", "application/json")

		handler.CreateOrder(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.Order
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, customer.ID, response.CustomerID)
		assert.Len(t, response.Products, 1)
	})

	t.Run("Invalid order data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			"POST",
			"/orders",
			strings.NewReader(`{"invalid":"json"`),
		)
		c.Request.Header.Add("Content-Type", "application/json")

		handler.CreateOrder(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateOrderStatus(t *testing.T) {
	db := setupTestDB()
	logger, _ := zap.NewDevelopment()
	handler := handlers.NewOrderHandler(db, logger)

	// Create test order
	customer := &models.Customer{Name: "Test Customer", Email: "test@example.com"}
	db.Create(customer)
	product := &models.Product{Name: "Book", Price: 29.99}
	db.Create(product)
	order := &models.Order{CustomerID: customer.ID, Products: []models.Product{*product}}
	db.Create(order)

	t.Run("Update valid order status", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprint(order.ID)}}
		c.Request = httptest.NewRequest(
			"PUT",
			fmt.Sprintf("/orders/%d/status", order.ID),
			strings.NewReader(`{"status":"shipped"}`),
		)
		c.Request.Header.Add("Content-Type", "application/json")

		handler.UpdateOrderStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid status data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprint(order.ID)}}
		c.Request = httptest.NewRequest(
			"PUT",
			fmt.Sprintf("/orders/%d/status", order.ID),
			strings.NewReader(`{"invalid":"json"`),
		)
		c.Request.Header.Add("Content-Type", "application/json")

		handler.UpdateOrderStatus(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetOrder(t *testing.T) {
	db := setupTestDB()
	logger, _ := zap.NewDevelopment()
	handler := handlers.NewOrderHandler(db, logger)

	// Create test order
	customer := &models.Customer{Name: "Test Customer", Email: "test@example.com"}
	db.Create(customer)
	product := &models.Product{Name: "Book", Price: 29.99}
	db.Create(product)
	order := &models.Order{CustomerID: customer.ID, Products: []models.Product{*product}}
	db.Create(order)

	t.Run("Get existing order", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: fmt.Sprint(order.ID)}}
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/orders/%d", order.ID), nil)

		handler.GetOrder(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.Order
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, order.ID, response.ID)
		assert.Equal(t, customer.ID, response.CustomerID)
	})

	t.Run("Get non-existent order", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest("GET", "/orders/999", nil)

		handler.GetOrder(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Invalid order ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "invalid"}}
		c.Request = httptest.NewRequest("GET", "/orders/invalid", nil)

		handler.GetOrder(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCreateProduct(t *testing.T) {
	db := setupTestDB()
	logger, _ := zap.NewDevelopment()
	handler := handlers.NewOrderHandler(db, logger)

	router := gin.Default()
	router.POST("/products", handler.CreateProduct)

	t.Run("Create valid product", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/products",
			strings.NewReader(`{"name":"Test Product","price":29.99}`))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Product
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Test Product", response.Name)
		assert.Equal(t, 29.99, response.Price)
	})

	t.Run("Invalid product data", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/products",
			strings.NewReader(`{"invalid":"json"`))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetProducts(t *testing.T) {
	db := setupTestDB()
	logger, _ := zap.NewDevelopment()
	handler := handlers.NewOrderHandler(db, logger)

	// Create test data
	db.Create(&models.Product{Name: "Product 1", Price: 10.0})
	db.Create(&models.Product{Name: "Product 2", Price: 20.0})

	router := gin.Default()
	router.GET("/products", handler.GetProducts)

	t.Run("Get all products", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var products []models.Product
		json.Unmarshal(w.Body.Bytes(), &products)
		
		assert.Len(t, products, 2)
		assert.Equal(t, "Product 1", products[0].Name)
		assert.Equal(t, 10.0, products[0].Price)
	})

	t.Run("Database error", func(t *testing.T) {
		// Drop the products table to simulate a DB error
		db.Migrator().DropTable(&models.Product{})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// Helper function for HTTP requests
func performRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}