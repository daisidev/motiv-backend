package repository

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
)

type EventRepository interface {
	CreateEvent(event *models.Event) error
	GetEventByID(id uuid.UUID) (*models.Event, error)
	GetEventsByHostID(hostID uuid.UUID) ([]*models.Event, error)
	GetAllEvents() ([]*models.Event, error)
	GetAllEventsWithPagination(page, limit int, search, tags, location, eventType, dateFrom, dateTo string) ([]*models.Event, int, error)
	UpdateEvent(event *models.Event) error
	DeleteEvent(id uuid.UUID) error
}
