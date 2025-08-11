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
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;"`
	EventID    uuid.UUID      `gorm:"type:uuid;not null"`
	Event      Event          `gorm:"foreignKey:EventID"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null"`
	User       User           `gorm:"foreignKey:UserID"`
	TicketID   uuid.UUID      `gorm:"type:uuid;not null"`
	Ticket     Ticket         `gorm:"foreignKey:TicketID"`
	Status     AttendeeStatus `gorm:"type:attendee_status;not null;default:'active'"`
	CheckedInAt *time.Time
	CheckedInBy *uuid.UUID `gorm:"type:uuid"` // Host who checked them in
}

func (a *Attendee) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New()
	return
}