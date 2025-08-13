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

// PaystackWebhookEvent represents the webhook payload from Paystack
type PaystackWebhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		ID                int64     `json:"id"`
		Domain            string    `json:"domain"`
		Status            string    `json:"status"`
		Reference         string    `json:"reference"`
		Amount            int64     `json:"amount"` // Amount in kobo
		Message           string    `json:"message"`
		GatewayResponse   string    `json:"gateway_response"`
		PaidAt            time.Time `json:"paid_at"`
		CreatedAt         time.Time `json:"created_at"`
		Channel           string    `json:"channel"`
		Currency          string    `json:"currency"`
		IPAddress         string    `json:"ip_address"`
		Metadata          struct {
			EventID      string `json:"eventId"`
			EventTitle   string `json:"eventTitle"`
			AttendeeData struct {
				FullName string `json:"fullName"`
				Email    string `json:"email"`
				Phone    string `json:"phone"`
			} `json:"attendeeData"`
			TicketDetails []struct {
				TicketTypeID   string  `json:"ticketTypeId"`
				TicketTypeName string  `json:"ticketTypeName"`
				Quantity       int     `json:"quantity"`
				Price          float64 `json:"price"`
			} `json:"ticketDetails"`
		} `json:"metadata"`
		Customer struct {
			ID           int64  `json:"id"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
			Phone        string `json:"phone"`
		} `json:"customer"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bin               string `json:"bin"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Channel           string `json:"channel"`
			CardType          string `json:"card_type"`
			Bank              string `json:"bank"`
			CountryCode       string `json:"country_code"`
			Brand             string `json:"brand"`
		} `json:"authorization"`
	} `json:"data"`
}

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