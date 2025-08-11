package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventView struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	EventID   uuid.UUID `gorm:"type:uuid;not null"`
	Event     Event     `gorm:"foreignKey:EventID"`
	UserID    *uuid.UUID `gorm:"type:uuid"` // nullable for anonymous views
	User      *User      `gorm:"foreignKey:UserID"`
	IPAddress string
	UserAgent string
	ViewedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type EventAnalytics struct {
	gorm.Model
	ID              uuid.UUID `gorm:"type:uuid;primary_key;"`
	EventID         uuid.UUID `gorm:"type:uuid;not null;unique"`
	Event           Event     `gorm:"foreignKey:EventID"`
	TotalViews      int       `gorm:"default:0"`
	UniqueViews     int       `gorm:"default:0"`
	TicketsSold     int       `gorm:"default:0"`
	Revenue         float64   `gorm:"default:0"`
	ConversionRate  float64   `gorm:"default:0"` // tickets sold / unique views
	WishlistAdds    int       `gorm:"default:0"`
	LastUpdated     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type HostAnalytics struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;primary_key;"`
	HostID         uuid.UUID `gorm:"type:uuid;not null;unique"`
	Host           User      `gorm:"foreignKey:HostID"`
	TotalEvents    int       `gorm:"default:0"`
	TotalRevenue   float64   `gorm:"default:0"`
	TotalAttendees int       `gorm:"default:0"`
	TotalViews     int       `gorm:"default:0"`
	AverageRating  float64   `gorm:"default:0"`
	LastUpdated    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (ev *EventView) BeforeCreate(tx *gorm.DB) (err error) {
	ev.ID = uuid.New()
	return
}

func (ea *EventAnalytics) BeforeCreate(tx *gorm.DB) (err error) {
	ea.ID = uuid.New()
	return
}

func (ha *HostAnalytics) BeforeCreate(tx *gorm.DB) (err error) {
	ha.ID = uuid.New()
	return
}