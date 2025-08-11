
package repository

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"gorm.io/gorm"
)

type eventRepoPG struct {
	db *gorm.DB
}

func NewEventRepoPG(db *gorm.DB) EventRepository {
	return &eventRepoPG{db}
}

func (r *eventRepoPG) CreateEvent(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepoPG) GetEventByID(id uuid.UUID) (*models.Event, error) {
	var event models.Event
	err := r.db.Preload("Host").Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepoPG) GetEventsByHostID(hostID uuid.UUID) ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.Preload("Host").Where("host_id = ?", hostID).Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *eventRepoPG) GetAllEvents() ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.Preload("Host").Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *eventRepoPG) UpdateEvent(event *models.Event) error {
	return r.db.Save(event).Error
}

func (r *eventRepoPG) DeleteEvent(id uuid.UUID) error {
	return r.db.Delete(&models.Event{}, id).Error
}
