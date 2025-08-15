package repository

import (
	"strings"

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
	err := r.db.Preload("Host").Preload("TicketTypes").Where("id = ?", id).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepoPG) GetEventsByHostID(hostID uuid.UUID) ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.Preload("Host").Preload("TicketTypes").Where("host_id = ?", hostID).Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *eventRepoPG) GetAllEvents() ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.Preload("Host").Preload("TicketTypes").Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *eventRepoPG) GetAllEventsWithPagination(page, limit int, search, tags, location, eventType, dateFrom, dateTo string) ([]*models.Event, int, error) {
	var events []*models.Event
	var total int64

	// Build the query
	query := r.db.Model(&models.Event{})

	// Apply filters
	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if tags != "" {
		tagList := strings.Split(tags, ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				query = query.Where("tags && ARRAY[?]", tag)
			}
		}
	}

	if location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}

	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	if dateFrom != "" {
		query = query.Where("start_date >= ?", dateFrom)
	}

	if dateTo != "" {
		query = query.Where("start_date <= ?", dateTo)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	err := query.Preload("Host").Preload("TicketTypes").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error

	if err != nil {
		return nil, 0, err
	}

	return events, int(total), nil
}

func (r *eventRepoPG) UpdateEvent(event *models.Event) error {
	// Use Select to only update specific fields, avoiding issues with host_id
	return r.db.Model(event).Select(
		"title", "description", "start_date", "start_time", "end_time", 
		"location", "latitude", "longitude", "place_id", "tags", "banner_image_url", "event_type", "status", "updated_at",
	).Updates(event).Error
}

func (r *eventRepoPG) GetSearchSuggestions(query string, limit int) ([]string, error) {
	var suggestions []string
	
	// Get unique event titles that match the query
	var titles []string
	err := r.db.Model(&models.Event{}).
		Select("DISTINCT title").
		Where("title ILIKE ?", "%"+query+"%").
		Limit(limit).
		Pluck("title", &titles).Error
	
	if err != nil {
		return nil, err
	}
	
	suggestions = append(suggestions, titles...)
	
	// If we don't have enough suggestions from titles, get from locations
	if len(suggestions) < limit {
		var locations []string
		remaining := limit - len(suggestions)
		err := r.db.Model(&models.Event{}).
			Select("DISTINCT location").
			Where("location ILIKE ? AND location NOT IN (?)", "%"+query+"%", suggestions).
			Limit(remaining).
			Pluck("location", &locations).Error
		
		if err != nil {
			return suggestions, nil // Return what we have so far
		}
		
		suggestions = append(suggestions, locations...)
	}
	
	return suggestions, nil
}

func (r *eventRepoPG) DeleteEvent(id uuid.UUID) error {
	return r.db.Delete(&models.Event{}, id).Error
}
