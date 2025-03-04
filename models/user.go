package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       string `gorm:"type:uuid;primary_key;"`
	Username string `gorm:"unique"`
	Email    string `gorm:"unique"`
	Role     UserRole
	Password string
}

func (b *User) BeforeCreate(tx *gorm.DB) (err error) {
	b.ID = uuid.New().String()
	return
}

type UserRole int

const (
	Admin UserRole = iota + 1
	Moderator
	Buddy
	Body
)
