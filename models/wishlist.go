
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wishlist struct {
	gorm.Model
	UserID  uuid.UUID `gorm:"type:uuid;primary_key"`
	EventID uuid.UUID `gorm:"type:uuid;primary_key"`
	User    User      `gorm:"foreignKey:UserID"`
	Event   Event     `gorm:"foreignKey:EventID"`
}
