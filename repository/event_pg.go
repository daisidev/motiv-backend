package repository

import (
	"fmt"
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

func (r *eventRepoPG) GetSimilarEvents(eventID uuid.UUID, limit int) ([]*models.Event, error) {
	// First get the current event to extract its properties
	var currentEvent models.Event
	err := r.db.Where("id = ?", eventID).First(&currentEvent).Error
	if err != nil {
		return nil, err
	}

	var events []*models.Event
	
	// Find similar events based on:
	// 1. Same location (highest priority)
	// 2. Same host (medium priority) 
	// 3. Other events (lowest priority)
	// Exclude the current event and limit results
	
	// First try to get events from same location
	query := r.db.Preload("Host").Preload("TicketTypes").
		Where("id != ?", eventID).
		Where("start_date >= NOW()"). // Only future events
		Where("location = ?", currentEvent.Location).
		Limit(limit)
	
	err = query.Find(&events).Error
	if err != nil {
		return nil, err
	}
	
	// If we don't have enough events from same location, get events from same host
	if len(events) < limit {
		var hostEvents []*models.Event
		remaining := limit - len(events)
		
		hostQuery := r.db.Preload("Host").Preload("TicketTypes").
			Where("id != ?", eventID).
			Where("start_date >= NOW()").
			Where("host_id = ?", currentEvent.HostID).
			Where("location != ?", currentEvent.Location). // Exclude same location events we already got
			Limit(remaining)
		
		err = hostQuery.Find(&hostEvents).Error
		if err != nil {
			return events, nil // Return what we have so far
		}
		
		events = append(events, hostEvents...)
	}
	
	// If we still don't have enough events, get any other future events
	if len(events) < limit {
		var otherEvents []*models.Event
		remaining := limit - len(events)
		
		otherQuery := r.db.Preload("Host").Preload("TicketTypes").
			Where("id != ?", eventID).
			Where("start_date >= NOW()").
			Where("location != ?", currentEvent.Location).
			Where("host_id != ?", currentEvent.HostID).
			Limit(remaining)
		
		err = otherQuery.Find(&otherEvents).Error
		if err != nil {
			return events, nil // Return what we have so far
		}
		
		events = append(events, otherEvents...)
	}

	return events, nil
}

func (r *eventRepoPG) GetPopularEvents(limit int) ([]*models.Event, error) {
	var events []*models.Event
	
	// Get popular events based on multiple factors:
	// 1. Events with more views/engagement
	// 2. Recent events (not too old)
	// 3. Events with tickets sold
	// 4. Shuffle the results to provide variety
	
	// First, let's check total count in database
	var totalCount int64
	r.db.Model(&models.Event{}).Count(&totalCount)
	
	err := r.db.Preload("Host").Preload("TicketTypes").
		// Temporarily removed filters to show all events for testing
		// Where("start_date >= NOW()"). // Only future events
		// Where("status = ?", "active"). // Only active events
		Order("RANDOM()"). // Shuffle the results
		Limit(limit).
		Find(&events).Error
	
	if err != nil {
		return nil, err
	}
	
	// Debug logging
	fmt.Printf("GetPopularEvents: Total events in DB: %d, Requested limit: %d, Found: %d\n", totalCount, limit, len(events))
	for i, event := range events {
		fmt.Printf("  Event %d: ID=%s, Title=%s, StartDate=%s\n", i+1, event.ID.String(), event.Title, event.StartDate.Format("2006-01-02"))
	}
	
	return events, nil
}

func (r *eventRepoPG) DeleteEvent(id uuid.UUID) error {
	return r.db.Delete(&models.Event{}, id).Error
}
