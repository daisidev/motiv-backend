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
	ID            uuid.UUID     `gorm:"type:uuid;primary_key;" json:"id"`
	TicketID      uuid.UUID     `gorm:"type:uuid;not null" json:"ticket_id"`
	Ticket        Ticket        `gorm:"foreignKey:TicketID" json:"ticket"`
	Amount        float64       `gorm:"not null" json:"amount"`
	Currency      string        `gorm:"default:'NGN'" json:"currency"`
	Status        PaymentStatus `gorm:"type:payment_status;not null" json:"status"`
	Method        PaymentMethod `gorm:"type:payment_method;not null" json:"method"`
	Reference     string        `gorm:"unique;not null" json:"reference"`
	ProcessedAt   *time.Time    `json:"processed_at"`
	FailureReason string        `json:"failure_reason"`
}

type Payout struct {
	gorm.Model
	ID          uuid.UUID     `gorm:"type:uuid;primary_key;" json:"id"`
	HostID      uuid.UUID     `gorm:"type:uuid;not null" json:"host_id"`
	Host        User          `gorm:"foreignKey:HostID" json:"host"`
	EventID     uuid.UUID     `gorm:"type:uuid;not null" json:"event_id"`
	Event       Event         `gorm:"foreignKey:EventID" json:"event"`
	Amount      float64       `gorm:"not null" json:"amount"`
	Currency    string        `gorm:"default:'NGN'" json:"currency"`
	Status      PaymentStatus `gorm:"type:payment_status;not null" json:"status"`
	Method      PaymentMethod `gorm:"type:payment_method;not null" json:"method"`
	Reference   string        `gorm:"unique;not null" json:"reference"`
	ProcessedAt *time.Time    `json:"processed_at"`
	PayoutDate  time.Time     `gorm:"not null" json:"payout_date"`
}

func (p *Payment) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}

func (p *Payout) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}