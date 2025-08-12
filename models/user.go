
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
	ID       uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name     string    `gorm:"not null" json:"name"`
	Username string    `gorm:"unique;not null" json:"username"`
	Email    string    `gorm:"unique;not null" json:"email"`
	Password string    `gorm:"not null" json:"-"` // Never serialize password
	Avatar   string    `json:"avatar"`
	Role     UserRole  `gorm:"type:varchar(20);not null;default:'guest'" json:"role"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}
