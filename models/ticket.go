
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;primary_key;"`
	EventID      uuid.UUID `gorm:"type:uuid;not null"`
	Event        Event     `gorm:"foreignKey:EventID"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	User         User      `gorm:"foreignKey:UserID"`
	TicketTypeID uuid.UUID `gorm:"type:uuid;not null"`
	TicketType   TicketType `gorm:"foreignKey:TicketTypeID"`
	QRCode       string
}

type TicketType struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;primary_key;"`
	EventID  uuid.UUID `gorm:"type:uuid;not null"`
	Event    Event     `gorm:"foreignKey:EventID"`
	Name     string    `gorm:"not null"`
	Price    float64   `gorm:"not null"`
	Quantity int       `gorm:"not null"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}

func (tt *TicketType) BeforeCreate(tx *gorm.DB) (err error) {
	tt.ID = uuid.New()
	return
}
