package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;primary_key;"`
	EventID  uuid.UUID `gorm:"type:uuid;not null"`
	Event    Event     `gorm:"foreignKey:EventID"`
	UserID   uuid.UUID `gorm:"type:uuid;not null"`
	User     User      `gorm:"foreignKey:UserID"`
	Rating   int       `gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment  string    `gorm:"type:text"`
	Helpful  int       `gorm:"default:0"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return
}