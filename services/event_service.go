package services

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
)

// EventQueryParams represents parameters for querying events
type EventQueryParams struct {
	Page      int
	Limit     int
	Search    string
	Tags      string
	Location  string
	EventType string
	DateFrom  string
	DateTo    string
}

// PaginatedEventResponse represents a paginated response for events
type PaginatedEventResponse struct {
	Data    []*models.Event `json:"data"`
	Total   int             `json:"total"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
	HasMore bool            `json:"hasMore"`
}

type EventService interface {
	CreateEvent(event *models.Event) error
	GetEventByID(id uuid.UUID) (*models.Event, error)
	GetEventsByHostID(hostID uuid.UUID) ([]*models.Event, error)
	GetAllEvents() ([]*models.Event, error)
	GetAllEventsWithPagination(params EventQueryParams) (*PaginatedEventResponse, error)
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

func (s *eventService) GetAllEventsWithPagination(params EventQueryParams) (*PaginatedEventResponse, error) {
	events, total, err := s.eventRepo.GetAllEventsWithPagination(params.Page, params.Limit, params.Search, params.Tags, params.Location, params.EventType, params.DateFrom, params.DateTo)
	if err != nil {
		return nil, err
	}

	hasMore := (params.Page * params.Limit) < total

	return &PaginatedEventResponse{
		Data:    events,
		Total:   total,
		Page:    params.Page,
		Limit:   params.Limit,
		HasMore: hasMore,
	}, nil
}

func (s *eventService) UpdateEvent(event *models.Event) error {
	return s.eventRepo.UpdateEvent(event)
}

func (s *eventService) DeleteEvent(id uuid.UUID) error {
	return s.eventRepo.DeleteEvent(id)
}
