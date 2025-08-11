
package repository

import (
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

func (r *ticketRepoPG) GetTicketsByUserID(userID uuid.UUID) ([]*models.Ticket, error) {
	var tickets []*models.Ticket
	err := r.db.Preload("Event").Preload("TicketType").Where("user_id = ?", userID).Find(&tickets).Error
	if err != nil {
		return nil, err
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
