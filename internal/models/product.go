// internal/models/product.go
package models

import "gorm.io/gorm"

// Product 模型对应数据库中的 'products' 表
type Product struct {
	gorm.Model          // GORM 提供的通用字段：ID, CreatedAt, UpdatedAt, DeletedAt
	Name        string  `gorm:"type:varchar(255);not null"`
	Description string  `gorm:"type:text"`
	Price       float64 `gorm:"type:decimal(10,2);not null"`
	Stock       int     `gorm:"type:int;not null;default:0"`
}
