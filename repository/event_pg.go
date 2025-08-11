
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

func (r *eventRepoPG) GetAllEventsWithPagination(page, limit int, search, tags, location string) ([]*models.Event, int, error) {
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
				query = query.Where("tags ILIKE ?", "%"+tag+"%")
			}
		}
	}

	if location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and get results
	offset := (page - 1) * limit
	err := query.Preload("Host").
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
	return r.db.Save(event).Error
}

func (r *eventRepoPG) DeleteEvent(id uuid.UUID) error {
	return r.db.Delete(&models.Event{}, id).Error
}
