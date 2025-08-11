
package services

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
)

type TicketService interface {
	PurchaseTicket(ticket *models.Ticket) error
	GetTicketsByUserID(userID uuid.UUID) ([]*models.Ticket, error)
	GetTicketByID(id uuid.UUID) (*models.Ticket, error)
}

type ticketService struct {
	ticketRepo repository.TicketRepository
}

func NewTicketService(ticketRepo repository.TicketRepository) TicketService {
	return &ticketService{ticketRepo}
}

func (s *ticketService) PurchaseTicket(ticket *models.Ticket) error {
	// In a real application, you would have more logic here, e.g., payment processing, QR code generation, etc.
	return s.ticketRepo.CreateTicket(ticket)
}

func (s *ticketService) GetTicketsByUserID(userID uuid.UUID) ([]*models.Ticket, error) {
	return s.ticketRepo.GetTicketsByUserID(userID)
}

func (s *ticketService) GetTicketByID(id uuid.UUID) (*models.Ticket, error) {
	return s.ticketRepo.GetTicketByID(id)
}
