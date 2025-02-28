package repository

import (
	"testing"
	"orderservice/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.Migrator().DropTable(&models.Order{}, &models.Customer{}, &models.Product{})
	db.AutoMigrate(&models.Order{}, &models.Customer{}, &models.Product{})
	return db
}

func TestOrderRepository(t *testing.T) {
	db := setupTestDB()
	repo := NewOrderRepository(db)

	t.Run("Create and get order", func(t *testing.T) {
		customer := &models.Customer{Name: "Test Customer", Email: "test@example.com"}
		err := repo.CreateCustomer(customer)
		assert.NoError(t, err)
		assert.NotZero(t, customer.ID)

		product := &models.Product{Name: "Test Product", Price: 29.99}
		err = repo.CreateProduct(product)
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)

		order := &models.Order{
			CustomerID: customer.ID,
			Products:   []models.Product{*product},
		}
		err = repo.CreateOrder(order)
		assert.NoError(t, err)
		assert.NotZero(t, order.ID)

		fetchedOrder, err := repo.GetOrder(order.ID)
		assert.NoError(t, err)
		assert.Equal(t, order.ID, fetchedOrder.ID)
		assert.Equal(t, customer.ID, fetchedOrder.CustomerID)
		assert.Len(t, fetchedOrder.Products, 1)
		assert.Equal(t, product.ID, fetchedOrder.Products[0].ID)
	})

	t.Run("Get non-existent order", func(t *testing.T) {
		_, err := repo.GetOrder(999)
		assert.Error(t, err)
	})

	t.Run("Update order status", func(t *testing.T) {
		customer := &models.Customer{Name: "Test Customer", Email: "test@example.com"}
		repo.CreateCustomer(customer)
		product := &models.Product{Name: "Test Product", Price: 29.99}
		repo.CreateProduct(product)
		order := &models.Order{CustomerID: customer.ID, Products: []models.Product{*product}}
		repo.CreateOrder(order)

		err := repo.UpdateOrderStatus(order.ID, "shipped")
		assert.NoError(t, err)

		updatedOrder, err := repo.GetOrder(order.ID)
		assert.NoError(t, err)
		assert.Equal(t, "shipped", updatedOrder.Status)
	})

	t.Run("Update non-existent order status", func(t *testing.T) {
		err := repo.UpdateOrderStatus(999, "shipped")
		assert.NoError(t, err)
	})
}

func TestCustomerRepository(t *testing.T) {
	db := setupTestDB()
	repo := NewOrderRepository(db)

	t.Run("Create customer", func(t *testing.T) {
		customer := &models.Customer{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		err := repo.CreateCustomer(customer)
		assert.NoError(t, err)
		assert.NotZero(t, customer.ID)

		var fetchedCustomer models.Customer
		err = db.First(&fetchedCustomer, customer.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, customer.Name, fetchedCustomer.Name)
		assert.Equal(t, customer.Email, fetchedCustomer.Email)
	})
}

func TestProductRepository(t *testing.T) {
	db := setupTestDB()
	repo := NewOrderRepository(db)

	t.Run("Create product", func(t *testing.T) {
		product := &models.Product{
			Name:  "Test Product",
			Price: 29.99,
		}
		err := repo.CreateProduct(product)
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)
	})

	t.Run("Get all products", func(t *testing.T) {
		products := []models.Product{
			{Name: "Product 1", Price: 10.0},
			{Name: "Product 2", Price: 20.0},
		}
		for _, p := range products {
			db.Create(&p)
		}

		var fetchedProducts []models.Product
		err := repo.GetAllProducts(&fetchedProducts)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(fetchedProducts), 2)
	})

	t.Run("Get all products with no products", func(t *testing.T) {
		db.Migrator().DropTable(&models.Product{})
		db.AutoMigrate(&models.Product{})

		var products []models.Product
		err := repo.GetAllProducts(&products)
		assert.NoError(t, err)
		assert.Empty(t, products)
	})
}