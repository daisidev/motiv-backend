
package repository

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
)

type TicketRepository interface {
	CreateTicket(ticket *models.Ticket) error
	UpdateTicket(ticket *models.Ticket) error
	GetTicketsByUserID(userID uuid.UUID) ([]*models.Ticket, error)
	GetTicketByID(id uuid.UUID) (*models.Ticket, error)
	GetByQRCode(qrCode string) (*models.Ticket, error)
	
	// Ticket Type methods
	CreateTicketType(ticketType *models.TicketType) error
	GetTicketTypesByEventID(eventID uuid.UUID) ([]*models.TicketType, error)
	GetTicketTypeByID(id uuid.UUID) (*models.TicketType, error)
	UpdateSoldQuantity(ticketTypeID uuid.UUID, quantity int) error
}
