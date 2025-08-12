package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wishlist struct {
	gorm.Model
	UserID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_event" json:"user_id"`
	EventID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_event" json:"event_id"`
	User    User      `gorm:"foreignKey:UserID" json:"user"`
	Event   Event     `gorm:"foreignKey:EventID" json:"event"`
}
