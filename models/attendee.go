package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendeeStatus string

const (
	AttendeeActive    AttendeeStatus = "active"
	AttendeeCheckedIn AttendeeStatus = "checked_in"
	AttendeeCancelled AttendeeStatus = "cancelled"
)

type Attendee struct {
	gorm.Model
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	EventID     uuid.UUID      `gorm:"type:uuid;not null" json:"event_id"`
	Event       Event          `gorm:"foreignKey:EventID" json:"event"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user"`
	TicketID    uuid.UUID      `gorm:"type:uuid;not null" json:"ticket_id"`
	Ticket      Ticket         `gorm:"foreignKey:TicketID" json:"ticket"`
	Status      AttendeeStatus `gorm:"type:attendee_status;not null;default:'active'" json:"status"`
	CheckedInAt *time.Time     `json:"checked_in_at"`
	CheckedInBy *uuid.UUID     `gorm:"type:uuid" json:"checked_in_by"` // Host who checked them in
}

func (a *Attendee) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New()
	return
}