
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	gorm.Model
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id"`
	EventID      uuid.UUID  `gorm:"type:uuid;not null" json:"event_id"`
	Event        Event      `gorm:"foreignKey:EventID" json:"event"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User         User       `gorm:"foreignKey:UserID" json:"user"`
	TicketTypeID uuid.UUID  `gorm:"type:uuid;not null" json:"ticket_type_id"`
	TicketType   TicketType `gorm:"foreignKey:TicketTypeID" json:"ticket_type"`
	QRCode       string     `json:"qr_code"`
	PaymentReference string `gorm:"not null" json:"payment_reference"`
	// Attendee information
	AttendeeFullName string `gorm:"not null" json:"attendee_full_name"`
	AttendeeEmail    string `gorm:"not null" json:"attendee_email"`
	AttendeePhone    string `gorm:"not null" json:"attendee_phone"`
	Quantity         int    `gorm:"not null;default:1" json:"quantity"`
}

type TicketType struct {
	gorm.Model
	ID            uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	EventID       uuid.UUID `gorm:"type:uuid;not null" json:"event_id"`
	Event         Event     `gorm:"foreignKey:EventID" json:"event"`
	Name          string    `gorm:"not null" json:"name"`
	Price         float64   `gorm:"not null" json:"price"`
	Description   string    `json:"description"`
	TotalQuantity int       `gorm:"not null;default:100" json:"total_quantity"`
	SoldQuantity  int       `gorm:"not null;default:0" json:"sold_quantity"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}

func (tt *TicketType) BeforeCreate(tx *gorm.DB) (err error) {
	tt.ID = uuid.New()
	return
}
