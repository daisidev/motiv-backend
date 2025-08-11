package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

type PaymentMethod string

const (
	BankTransfer PaymentMethod = "bank_transfer"
	Card         PaymentMethod = "card"
	Wallet       PaymentMethod = "wallet"
)

type Payment struct {
	gorm.Model
	ID            uuid.UUID     `gorm:"type:uuid;primary_key;"`
	TicketID      uuid.UUID     `gorm:"type:uuid;not null"`
	Ticket        Ticket        `gorm:"foreignKey:TicketID"`
	Amount        float64       `gorm:"not null"`
	Currency      string        `gorm:"default:'NGN'"`
	Status        PaymentStatus `gorm:"type:payment_status;not null"`
	Method        PaymentMethod `gorm:"type:payment_method;not null"`
	Reference     string        `gorm:"unique;not null"`
	ProcessedAt   *time.Time
	FailureReason string
}

type Payout struct {
	gorm.Model
	ID          uuid.UUID     `gorm:"type:uuid;primary_key;"`
	HostID      uuid.UUID     `gorm:"type:uuid;not null"`
	Host        User          `gorm:"foreignKey:HostID"`
	EventID     uuid.UUID     `gorm:"type:uuid;not null"`
	Event       Event         `gorm:"foreignKey:EventID"`
	Amount      float64       `gorm:"not null"`
	Currency    string        `gorm:"default:'NGN'"`
	Status      PaymentStatus `gorm:"type:payment_status;not null"`
	Method      PaymentMethod `gorm:"type:payment_method;not null"`
	Reference   string        `gorm:"unique;not null"`
	ProcessedAt *time.Time
	PayoutDate  time.Time `gorm:"not null"`
}

func (p *Payment) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}

func (p *Payout) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}