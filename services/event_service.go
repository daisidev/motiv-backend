
package services

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
)

type EventService interface {
	CreateEvent(event *models.Event) error
	GetEventByID(id uuid.UUID) (*models.Event, error)
	GetEventsByHostID(hostID uuid.UUID) ([]*models.Event, error)
	GetAllEvents() ([]*models.Event, error)
	UpdateEvent(event *models.Event) error
	DeleteEvent(id uuid.UUID) error
}

type eventService struct {
	eventRepo repository.EventRepository
}

func NewEventService(eventRepo repository.EventRepository) EventService {
	return &eventService{eventRepo}
}

func (s *eventService) CreateEvent(event *models.Event) error {
	return s.eventRepo.CreateEvent(event)
}

func (s *eventService) GetEventByID(id uuid.UUID) (*models.Event, error) {
	return s.eventRepo.GetEventByID(id)
}

func (s *eventService) GetEventsByHostID(hostID uuid.UUID) ([]*models.Event, error) {
	return s.eventRepo.GetEventsByHostID(hostID)
}

func (s *eventService) GetAllEvents() ([]*models.Event, error) {
	return s.eventRepo.GetAllEvents()
}

func (s *eventService) UpdateEvent(event *models.Event) error {
	return s.eventRepo.UpdateEvent(event)
}

func (s *eventService) DeleteEvent(id uuid.UUID) error {
	return s.eventRepo.DeleteEvent(id)
}
