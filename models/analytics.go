package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventView struct {
	gorm.Model
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id"`
	EventID   uuid.UUID  `gorm:"type:uuid;not null" json:"event_id"`
	Event     Event      `gorm:"foreignKey:EventID" json:"event"`
	UserID    *uuid.UUID `gorm:"type:uuid" json:"user_id"` // nullable for anonymous views
	User      *User      `gorm:"foreignKey:UserID" json:"user"`
	IPAddress string     `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	ViewedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"viewed_at"`
}

type EventAnalytics struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	EventID        uuid.UUID `gorm:"type:uuid;not null;unique" json:"event_id"`
	Event          Event     `gorm:"foreignKey:EventID" json:"event"`
	TotalViews     int       `gorm:"default:0" json:"total_views"`
	UniqueViews    int       `gorm:"default:0" json:"unique_views"`
	TicketsSold    int       `gorm:"default:0" json:"tickets_sold"`
	Revenue        float64   `gorm:"default:0" json:"revenue"`
	ConversionRate float64   `gorm:"default:0" json:"conversion_rate"` // tickets sold / unique views
	WishlistAdds   int       `gorm:"default:0" json:"wishlist_adds"`
	LastUpdated    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"last_updated"`
}

type HostAnalytics struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	HostID         uuid.UUID `gorm:"type:uuid;not null;unique" json:"host_id"`
	Host           User      `gorm:"foreignKey:HostID" json:"host"`
	TotalEvents    int       `gorm:"default:0" json:"total_events"`
	TotalRevenue   float64   `gorm:"default:0" json:"total_revenue"`
	TotalAttendees int       `gorm:"default:0" json:"total_attendees"`
	TotalViews     int       `gorm:"default:0" json:"total_views"`
	AverageRating  float64   `gorm:"default:0" json:"average_rating"`
	LastUpdated    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"last_updated"`
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