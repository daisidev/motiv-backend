
package services

import (
	"fmt"
	
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
)

type TicketService interface {
	PurchaseTicket(ticket *models.Ticket) error
	CreateTicketWithQR(ticket *models.Ticket) error
	GetTicketsByUserID(userID uuid.UUID) ([]*models.Ticket, error)
	GetTicketByID(id uuid.UUID) (*models.Ticket, error)
	
	// Ticket Type methods
	CreateTicketType(ticketType *models.TicketType) error
	GetTicketTypesByEventID(eventID uuid.UUID) ([]*models.TicketType, error)
	GetTicketTypeByID(ticketTypeID uuid.UUID) (*models.TicketType, error)
	UpdateSoldQuantity(ticketTypeID uuid.UUID, quantity int) error
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

func (s *ticketService) CreateTicketWithQR(ticket *models.Ticket) error {
	// First create the ticket to get the ID
	err := s.ticketRepo.CreateTicket(ticket)
	if err != nil {
		return err
	}
	
	// Now generate QR code data with the actual ticket ID
	qrData := fmt.Sprintf("MOTIV-TICKET:%s:%s:%s", ticket.ID.String(), ticket.EventID.String(), ticket.UserID.String())
	ticket.QRCode = qrData
	
	// Update the ticket with the QR code
	return s.ticketRepo.UpdateTicket(ticket)
}

func (s *ticketService) GetTicketTypeByID(ticketTypeID uuid.UUID) (*models.TicketType, error) {
	return s.ticketRepo.GetTicketTypeByID(ticketTypeID)
}

func (s *ticketService) UpdateSoldQuantity(ticketTypeID uuid.UUID, quantity int) error {
	return s.ticketRepo.UpdateSoldQuantity(ticketTypeID, quantity)
}

func (s *ticketService) GetTicketsByUserID(userID uuid.UUID) ([]*models.Ticket, error) {
	return s.ticketRepo.GetTicketsByUserID(userID)
}

func (s *ticketService) GetTicketByID(id uuid.UUID) (*models.Ticket, error) {
	return s.ticketRepo.GetTicketByID(id)
}

// Ticket Type methods
func (s *ticketService) CreateTicketType(ticketType *models.TicketType) error {
	return s.ticketRepo.CreateTicketType(ticketType)
}

func (s *ticketService) GetTicketTypesByEventID(eventID uuid.UUID) ([]*models.TicketType, error) {
	return s.ticketRepo.GetTicketTypesByEventID(eventID)
}
