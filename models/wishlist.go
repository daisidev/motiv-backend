
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wishlist struct {
	gorm.Model
	UserID  uuid.UUID `gorm:"type:uuid;primary_key" json:"user_id"`
	EventID uuid.UUID `gorm:"type:uuid;primary_key" json:"event_id"`
	User    User      `gorm:"foreignKey:UserID" json:"user"`
	Event   Event     `gorm:"foreignKey:EventID" json:"event"`
}
