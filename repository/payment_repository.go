package repository

import (
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	// Payments
	CreatePayment(payment *models.Payment) error
	GetPaymentByID(id uuid.UUID) (*models.Payment, error)
	GetPaymentByReference(reference string) (*models.Payment, error)
	UpdatePayment(payment *models.Payment) error
	GetPaymentsByTicketID(ticketID uuid.UUID) ([]models.Payment, error)
	
	// Payouts
	CreatePayout(payout *models.Payout) error
	GetPayoutByID(id uuid.UUID) (*models.Payout, error)
	GetPayoutsByHostID(hostID uuid.UUID, limit, offset int) ([]models.Payout, error)
	UpdatePayout(payout *models.Payout) error
	GetPendingPayouts(hostID uuid.UUID) ([]models.Payout, error)
	
	// Financial Stats
	GetHostEarnings(hostID uuid.UUID) (float64, error)
	GetHostMonthlyEarnings(hostID uuid.UUID, year, month int) (float64, error)
	GetEventRevenue(eventID uuid.UUID) (float64, error)
}

type paymentRepoPG struct {
	db *gorm.DB
}

func NewPaymentRepoPG(db *gorm.DB) PaymentRepository {
	return &paymentRepoPG{db: db}
}

// Payment methods
func (p *paymentRepoPG) CreatePayment(payment *models.Payment) error {
	return p.db.Create(payment).Error
}

func (p *paymentRepoPG) GetPaymentByID(id uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := p.db.Preload("Ticket").First(&payment, "id = ?", id).Error
	return &payment, err
}

func (p *paymentRepoPG) GetPaymentByReference(reference string) (*models.Payment, error) {
	var payment models.Payment
	err := p.db.Preload("Ticket").First(&payment, "reference = ?", reference).Error
	return &payment, err
}

func (p *paymentRepoPG) UpdatePayment(payment *models.Payment) error {
	return p.db.Save(payment).Error
}

func (p *paymentRepoPG) GetPaymentsByTicketID(ticketID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	err := p.db.Where("ticket_id = ?", ticketID).Find(&payments).Error
	return payments, err
}

// Payout methods
func (p *paymentRepoPG) CreatePayout(payout *models.Payout) error {
	return p.db.Create(payout).Error
}

func (p *paymentRepoPG) GetPayoutByID(id uuid.UUID) (*models.Payout, error) {
	var payout models.Payout
	err := p.db.Preload("Host").Preload("Event").First(&payout, "id = ?", id).Error
	return &payout, err
}

func (p *paymentRepoPG) GetPayoutsByHostID(hostID uuid.UUID, limit, offset int) ([]models.Payout, error) {
	var payouts []models.Payout
	err := p.db.Preload("Event").Where("host_id = ?", hostID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&payouts).Error
	return payouts, err
}

func (p *paymentRepoPG) UpdatePayout(payout *models.Payout) error {
	return p.db.Save(payout).Error
}

func (p *paymentRepoPG) GetPendingPayouts(hostID uuid.UUID) ([]models.Payout, error) {
	var payouts []models.Payout
	err := p.db.Preload("Event").
		Where("host_id = ? AND status = ?", hostID, models.PaymentPending).
		Find(&payouts).Error
	return payouts, err
}

// Financial Stats methods
func (p *paymentRepoPG) GetHostEarnings(hostID uuid.UUID) (float64, error) {
	var totalEarnings float64
	err := p.db.Model(&models.Payment{}).
		Joins("JOIN tickets ON payments.ticket_id = tickets.id").
		Joins("JOIN events ON tickets.event_id = events.id").
		Where("events.host_id = ? AND payments.status = ?", hostID, models.PaymentCompleted).
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&totalEarnings).Error
	return totalEarnings, err
}

func (p *paymentRepoPG) GetHostMonthlyEarnings(hostID uuid.UUID, year, month int) (float64, error) {
	var monthlyEarnings float64
	err := p.db.Model(&models.Payment{}).
		Joins("JOIN tickets ON payments.ticket_id = tickets.id").
		Joins("JOIN events ON tickets.event_id = events.id").
		Where("events.host_id = ? AND payments.status = ? AND EXTRACT(YEAR FROM payments.created_at) = ? AND EXTRACT(MONTH FROM payments.created_at) = ?", 
			hostID, models.PaymentCompleted, year, month).
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&monthlyEarnings).Error
	return monthlyEarnings, err
}

func (p *paymentRepoPG) GetEventRevenue(eventID uuid.UUID) (float64, error) {
	var revenue float64
	err := p.db.Model(&models.Payment{}).
		Joins("JOIN tickets ON payments.ticket_id = tickets.id").
		Where("tickets.event_id = ? AND payments.status = ?", eventID, models.PaymentCompleted).
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&revenue).Error
	return revenue, err
}