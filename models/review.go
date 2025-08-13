package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Review struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	EventID   uuid.UUID  `gorm:"type:uuid;not null" json:"event_id"`
	Event     Event      `gorm:"foreignKey:EventID" json:"event"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID" json:"user"`
	Rating    int        `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment   string     `gorm:"type:text" json:"comment"`
	Helpful   int        `gorm:"default:0" json:"helpful"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return
}