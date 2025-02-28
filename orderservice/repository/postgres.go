package repository

import (
	"orderservice/models"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) GetOrder(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Products").First(&order, id).Error
	return &order, err
}

func (r *OrderRepository) UpdateOrderStatus(id uint, status string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *OrderRepository) CreateCustomer(customer *models.Customer) error {
	return r.db.Create(customer).Error
}

func (r *OrderRepository) CreateProduct(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *OrderRepository) GetAllProducts(products *[]models.Product) error {
	return r.db.Find(products).Error
}