
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	GuestRole     UserRole = "guest"
	HostRole      UserRole = "host"
	AdminRole     UserRole = "admin"
	SuperhostRole UserRole = "superhost"
)

type User struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name     string    `gorm:"not null"`
	Email    string    `gorm:"unique;not null"`
	Password string    `gorm:"not null"`
	Avatar   string
	Role     UserRole `gorm:"type:varchar(20);not null;default:'guest'"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}
