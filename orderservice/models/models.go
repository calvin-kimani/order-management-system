package models

import (
	"gorm.io/gorm"
)

type Customer struct {
	gorm.Model
	Name  string `gorm:"not null" json:"name"`
	Email string `gorm:"unique;not null" json:"email"`
}

type Product struct {
	gorm.Model
	Name  string  `gorm:"not null" json:"name"`
	Price float64 `gorm:"not null" json:"price"`
}

type Order struct {
	gorm.Model
	CustomerID uint      `gorm:"not null" json:"customer_id"`
	Products   []Product `gorm:"many2many:order_products;" json:"products"`
	Status     string    `gorm:"default:'pending'" json:"status"`
	Total      float64   `gorm:"-" json:"total"` // Virtual field
}

// Hooks
func (o *Order) AfterFind(tx *gorm.DB) (err error) {
	for _, product := range o.Products {
		o.Total += product.Price
	}
	return nil
}