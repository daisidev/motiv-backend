
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
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;"`
	Title          string         `gorm:"not null"`
	Description    string
	StartDate      time.Time      `gorm:"not null"`
	StartTime      string         `gorm:"not null"`
	EndTime        string         `gorm:"not null"`
	Location       string         `gorm:"not null"`
	Tags           pq.StringArray `gorm:"type:text[]"`
	BannerImageURL string
	HostID         uuid.UUID      `gorm:"type:uuid;not null"`
	Host           User           `gorm:"foreignKey:HostID"`
	Status         EventStatus    `gorm:"type:varchar(20);not null;default:'draft'"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.New()
	return
}
