// internal/models/user.go
package models

import (
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID   string  `gorm:"type:varchar(36);unique;not null;default:'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx'"`
	Name     string  `gorm:"type:varchar(255);not null"`
	Email    *string `gorm:"type:varchar(225);unique;null"`
	Phone    *string `gorm:"type:varchar(225);unique;null"`
	Password string  `gorm:"type:varchar(256);not null"`
}
type UserJWT struct {
	UserID string
	jwt.RegisteredClaims
}
type UserLogin struct {
	Loginfo  string
	Password string
}
