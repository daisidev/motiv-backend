package services

import (
	"fmt"
	"time"

	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
	"github.com/google/uuid"
)

type PaymentService interface {
	// Payment processing
	ProcessPayment(ticketID uuid.UUID, amount float64, method models.PaymentMethod) (*models.Payment, error)
	UpdatePaymentStatus(reference string, status models.PaymentStatus, failureReason string) error
	GetPaymentByReference(reference string) (*models.Payment, error)
	
	// Payouts
	CreatePayout(hostID, eventID uuid.UUID, amount float64) (*models.Payout, error)
	GetHostPayouts(hostID uuid.UUID, page, limit int) ([]models.Payout, error)
	ProcessPayout(payoutID uuid.UUID) error
	GetPendingPayouts(hostID uuid.UUID) ([]models.Payout, error)
	
	// Financial stats
	GetHostEarnings(hostID uuid.UUID) (map[string]interface{}, error)
	GetEventRevenue(eventID uuid.UUID) (float64, error)
}

type paymentService struct {
	paymentRepo repository.PaymentRepository
}

func NewPaymentService(paymentRepo repository.PaymentRepository) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
	}
}

func (s *paymentService) ProcessPayment(ticketID uuid.UUID, amount float64, method models.PaymentMethod) (*models.Payment, error) {
	// Generate unique reference
	reference := fmt.Sprintf("PAY-%s-%d", ticketID.String()[:8], time.Now().Unix())
	
	payment := &models.Payment{
		TicketID:  ticketID,
		Amount:    amount,
		Currency:  "NGN",
		Status:    models.PaymentPending,
		Method:    method,
		Reference: reference,
	}
	
	err := s.paymentRepo.CreatePayment(payment)
	if err != nil {
		return nil, err
	}
	
	// Here you would integrate with actual payment processor
	// For now, we'll simulate successful payment
	payment.Status = models.PaymentCompleted
	now := time.Now()
	payment.ProcessedAt = &now
	
	err = s.paymentRepo.UpdatePayment(payment)
	return payment, err
}

func (s *paymentService) UpdatePaymentStatus(reference string, status models.PaymentStatus, failureReason string) error {
	payment, err := s.paymentRepo.GetPaymentByReference(reference)
	if err != nil {
		return err
	}
	
	payment.Status = status
	if status == models.PaymentCompleted {
		now := time.Now()
		payment.ProcessedAt = &now
	} else if status == models.PaymentFailed {
		payment.FailureReason = failureReason
	}
	
	return s.paymentRepo.UpdatePayment(payment)
}

func (s *paymentService) GetPaymentByReference(reference string) (*models.Payment, error) {
	return s.paymentRepo.GetPaymentByReference(reference)
}

func (s *paymentService) CreatePayout(hostID, eventID uuid.UUID, amount float64) (*models.Payout, error) {
	// Generate unique reference
	reference := fmt.Sprintf("PAYOUT-%s-%d", hostID.String()[:8], time.Now().Unix())
	
	// Calculate payout date (5 business days after event)
	payoutDate := time.Now().AddDate(0, 0, 7) // 7 days for simplicity
	
	payout := &models.Payout{
		HostID:     hostID,
		EventID:    eventID,
		Amount:     amount * 0.99, // 1% commission
		Currency:   "NGN",
		Status:     models.PaymentPending,
		Method:     models.BankTransfer,
		Reference:  reference,
		PayoutDate: payoutDate,
	}
	
	return payout, s.paymentRepo.CreatePayout(payout)
}

func (s *paymentService) GetHostPayouts(hostID uuid.UUID, page, limit int) ([]models.Payout, error) {
	offset := (page - 1) * limit
	return s.paymentRepo.GetPayoutsByHostID(hostID, limit, offset)
}

func (s *paymentService) ProcessPayout(payoutID uuid.UUID) error {
	payout, err := s.paymentRepo.GetPayoutByID(payoutID)
	if err != nil {
		return err
	}
	
	// Here you would integrate with actual payout processor
	// For now, we'll simulate successful payout
	payout.Status = models.PaymentCompleted
	now := time.Now()
	payout.ProcessedAt = &now
	
	return s.paymentRepo.UpdatePayout(payout)
}

func (s *paymentService) GetPendingPayouts(hostID uuid.UUID) ([]models.Payout, error) {
	return s.paymentRepo.GetPendingPayouts(hostID)
}

func (s *paymentService) GetHostEarnings(hostID uuid.UUID) (map[string]interface{}, error) {
	totalEarnings, err := s.paymentRepo.GetHostEarnings(hostID)
	if err != nil {
		return nil, err
	}
	
	// Get current month earnings
	now := time.Now()
	monthlyEarnings, err := s.paymentRepo.GetHostMonthlyEarnings(hostID, now.Year(), int(now.Month()))
	if err != nil {
		return nil, err
	}
	
	// Get pending payouts
	pendingPayouts, err := s.paymentRepo.GetPendingPayouts(hostID)
	if err != nil {
		return nil, err
	}
	
	var pendingAmount float64
	for _, payout := range pendingPayouts {
		pendingAmount += payout.Amount
	}
	
	return map[string]interface{}{
		"total_earnings":    totalEarnings,
		"monthly_earnings":  monthlyEarnings,
		"pending_payouts":   pendingAmount,
		"next_payout_date":  time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
	}, nil
}

func (s *paymentService) GetEventRevenue(eventID uuid.UUID) (float64, error) {
	return s.paymentRepo.GetEventRevenue(eventID)
}