package repository

import (
	"log"

	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"gorm.io/gorm"
)

type ticketRepoPG struct {
	db *gorm.DB
}

func NewTicketRepoPG(db *gorm.DB) TicketRepository {
	return &ticketRepoPG{db}
}

func (r *ticketRepoPG) CreateTicket(ticket *models.Ticket) error {
	return r.db.Create(ticket).Error
}

func (r *ticketRepoPG) UpdateTicket(ticket *models.Ticket) error {
	return r.db.Save(ticket).Error
}

func (r *ticketRepoPG) GetTicketsByUserID(userID uuid.UUID) ([]*models.Ticket, error) {
	var tickets []*models.Ticket
	err := r.db.Preload("Event").Preload("TicketType").Where("user_id = ?", userID).Find(&tickets).Error
	if err != nil {
		return nil, err
	}

	// Debug logging
	for _, ticket := range tickets {
		log.Printf("Ticket ID: %s, Event ID: %s, Event Title: %s",
			ticket.ID.String(),
			ticket.EventID.String(),
			func() string {
				if ticket.Event.Title != "" {
					return ticket.Event.Title
				}
				return "NO_TITLE"
			}())
	}

	return tickets, nil
}

func (r *ticketRepoPG) GetTicketByID(id uuid.UUID) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Preload("Event").Preload("TicketType").Where("id = ?", id).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// Ticket Type methods
func (r *ticketRepoPG) CreateTicketType(ticketType *models.TicketType) error {
	return r.db.Create(ticketType).Error
}

func (r *ticketRepoPG) GetTicketTypesByEventID(eventID uuid.UUID) ([]*models.TicketType, error) {
	var ticketTypes []*models.TicketType
	err := r.db.Where("event_id = ?", eventID).Find(&ticketTypes).Error
	if err != nil {
		return nil, err
	}
	return ticketTypes, nil
}

func (r *ticketRepoPG) GetTicketTypeByID(id uuid.UUID) (*models.TicketType, error) {
	var ticketType models.TicketType
	err := r.db.Where("id = ?", id).First(&ticketType).Error
	if err != nil {
		return nil, err
	}
	return &ticketType, nil
}

func (r *ticketRepoPG) UpdateSoldQuantity(ticketTypeID uuid.UUID, quantity int) error {
	return r.db.Model(&models.TicketType{}).
		Where("id = ?", ticketTypeID).
		Update("sold_quantity", gorm.Expr("sold_quantity + ?", quantity)).Error
}
