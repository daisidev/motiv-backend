package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type EventStatus string

const (
	DraftEvent     EventStatus = "draft"
	ActiveEvent    EventStatus = "active"
	CancelledEvent EventStatus = "cancelled"
)

type Event struct {
	gorm.Model
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	Title          string         `gorm:"not null" json:"title"`
	Description    string         `json:"description"`
	StartDate      time.Time      `gorm:"not null" json:"start_date"`
	StartTime      string         `gorm:"not null" json:"start_time"`
	EndTime        string         `gorm:"not null" json:"end_time"`
	Location       string         `gorm:"not null" json:"location"`
	Latitude       *float64       `json:"latitude,omitempty"`
	Longitude      *float64       `json:"longitude,omitempty"`
	PlaceID        *string        `json:"place_id,omitempty"`
	Tags           pq.StringArray `gorm:"type:text[]" json:"tags"`
	BannerImageURL string         `json:"banner_image_url"`
	EventType      string         `gorm:"type:varchar(20);not null;default:'ticketed'" json:"event_type"` // "ticketed" or "free"
	HostID         uuid.UUID      `gorm:"type:uuid;not null" json:"host_id"`
	Host           User           `gorm:"foreignKey:HostID" json:"host"`
	TicketTypes    []TicketType   `gorm:"foreignKey:EventID" json:"ticket_types,omitempty"`
	Status         EventStatus    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.New()
	return
}
