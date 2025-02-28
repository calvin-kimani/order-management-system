package models

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Order{}, &Customer{}, &Product{})
	return db
}

func TestCustomerModel(t *testing.T) {
	db := setupTestDB()

	t.Run("Create customer with valid data", func(t *testing.T) {
		customer := &Customer{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		err := db.Create(customer).Error
		assert.NoError(t, err)
		assert.NotZero(t, customer.ID)
	})

	t.Run("Create customer with duplicate email", func(t *testing.T) {
		customer1 := &Customer{
			Name:  "John Doe",
			Email: "duplicate@example.com",
		}
		customer2 := &Customer{
			Name:  "Jane Doe",
			Email: "duplicate@example.com",
		}
		err := db.Create(customer1).Error
		assert.NoError(t, err)
		err = db.Create(customer2).Error
		assert.Error(t, err)
	})
}

func TestProductModel(t *testing.T) {
	db := setupTestDB()

	t.Run("Create product with valid data", func(t *testing.T) {
		product := &Product{
			Name:  "Test Product",
			Price: 29.99,
		}
		err := db.Create(product).Error
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)
	})

	t.Run("Create product with zero price", func(t *testing.T) {
		product := &Product{
			Name:  "Free Product",
			Price: 0,
		}
		err := db.Create(product).Error
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)
	})
}

func TestOrderModel(t *testing.T) {
	db := setupTestDB()

	customer := &Customer{Name: "Test Customer", Email: "test@example.com"}
	db.Create(customer)

	products := []Product{
		{Name: "Product 1", Price: 10.0},
		{Name: "Product 2", Price: 20.0},
	}
	for i := range products {
		db.Create(&products[i])
	}

	t.Run("Create order and calculate total", func(t *testing.T) {
		order := &Order{
			CustomerID: customer.ID,
			Products:   products,
		}
		err := db.Create(order).Error
		assert.NoError(t, err)
		assert.NotZero(t, order.ID)

		var fetchedOrder Order
		err = db.Preload("Products").First(&fetchedOrder, order.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, 30.0, fetchedOrder.Total)
	})

	t.Run("Create order without products", func(t *testing.T) {
		order := &Order{
			CustomerID: customer.ID,
			Products:   []Product{},
		}
		err := db.Create(order).Error
		assert.NoError(t, err)
		assert.NotZero(t, order.ID)

		var fetchedOrder Order
		err = db.Preload("Products").First(&fetchedOrder, order.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, 0.0, fetchedOrder.Total)
	})

	t.Run("Update order status", func(t *testing.T) {
		order := &Order{
			CustomerID: customer.ID,
			Products:   products,
			Status:     "pending",
		}
		err := db.Create(order).Error
		assert.NoError(t, err)

		err = db.Model(order).Update("status", "shipped").Error
		assert.NoError(t, err)

		var fetchedOrder Order
		err = db.First(&fetchedOrder, order.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "shipped", fetchedOrder.Status)
	})
}
