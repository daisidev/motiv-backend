package repository

import (
	"strings"

	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"gorm.io/gorm"
)

type AttendeeRepository interface {
	Create(attendee *models.Attendee) error
	GetByID(id uuid.UUID) (*models.Attendee, error)
	GetByEventID(eventID uuid.UUID, limit, offset int) ([]models.Attendee, error)
	GetEventAttendeesTotalCount(eventID uuid.UUID) (int64, error)
	GetByHostID(hostID uuid.UUID, limit, offset int) ([]models.Attendee, error)
	GetHostAttendeesTotalCount(hostID uuid.UUID) (int64, error)
	GetByHostIDWithFilters(hostID uuid.UUID, limit, offset int, eventID *uuid.UUID, ticketType, status, search string) ([]models.Attendee, error)
	GetHostAttendeesTotalCountWithFilters(hostID uuid.UUID, eventID *uuid.UUID, ticketType, status, search string) (int64, error)
	Update(attendee *models.Attendee) error
	Delete(id uuid.UUID) error
	CheckInAttendee(attendeeID, checkedInBy uuid.UUID) error
	GetEventAttendeeStats(eventID uuid.UUID) (map[string]int64, error)
	GetHostAttendeeStats(hostID uuid.UUID) (map[string]int64, error)
}

type attendeeRepoPG struct {
	db *gorm.DB
}

func NewAttendeeRepoPG(db *gorm.DB) AttendeeRepository {
	return &attendeeRepoPG{db: db}
}

func (a *attendeeRepoPG) Create(attendee *models.Attendee) error {
	return a.db.Create(attendee).Error
}

func (a *attendeeRepoPG) GetByID(id uuid.UUID) (*models.Attendee, error) {
	var attendee models.Attendee
	err := a.db.Preload("User").Preload("Event").Preload("Ticket").
		First(&attendee, "id = ?", id).Error
	return &attendee, err
}

func (a *attendeeRepoPG) GetByEventID(eventID uuid.UUID, limit, offset int) ([]models.Attendee, error) {
	var attendees []models.Attendee
	err := a.db.Preload("User").Preload("Ticket").Preload("Ticket.TicketType").
		Where("event_id = ?", eventID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&attendees).Error
	return attendees, err
}

func (a *attendeeRepoPG) GetEventAttendeesTotalCount(eventID uuid.UUID) (int64, error) {
	var count int64
	err := a.db.Model(&models.Attendee{}).
		Where("event_id = ?", eventID).
		Count(&count).Error
	return count, err
}

func (a *attendeeRepoPG) GetByHostID(hostID uuid.UUID, limit, offset int) ([]models.Attendee, error) {
	var attendees []models.Attendee
	err := a.db.Preload("User").Preload("Event").Preload("Ticket").Preload("Ticket.TicketType").
		Joins("JOIN events ON attendees.event_id = events.id").
		Where("events.host_id = ?", hostID).
		Order("attendees.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&attendees).Error
	return attendees, err
}

func (a *attendeeRepoPG) GetHostAttendeesTotalCount(hostID uuid.UUID) (int64, error) {
	var count int64
	err := a.db.Model(&models.Attendee{}).
		Joins("JOIN events ON attendees.event_id = events.id").
		Where("events.host_id = ?", hostID).
		Count(&count).Error
	return count, err
}

func (a *attendeeRepoPG) Update(attendee *models.Attendee) error {
	return a.db.Save(attendee).Error
}

func (a *attendeeRepoPG) Delete(id uuid.UUID) error {
	return a.db.Delete(&models.Attendee{}, "id = ?", id).Error
}

func (a *attendeeRepoPG) CheckInAttendee(attendeeID, checkedInBy uuid.UUID) error {
	return a.db.Model(&models.Attendee{}).
		Where("id = ?", attendeeID).
		Updates(map[string]interface{}{
			"status":        models.AttendeeCheckedIn,
			"checked_in_at": gorm.Expr("NOW()"),
			"checked_in_by": checkedInBy,
		}).Error
}

func (a *attendeeRepoPG) GetEventAttendeeStats(eventID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total attendees
	var total int64
	a.db.Model(&models.Attendee{}).Where("event_id = ?", eventID).Count(&total)
	stats["total"] = total

	// Checked in
	var checkedIn int64
	a.db.Model(&models.Attendee{}).
		Where("event_id = ? AND status = ?", eventID, models.AttendeeCheckedIn).
		Count(&checkedIn)
	stats["checked_in"] = checkedIn

	// Active (not checked in)
	var active int64
	a.db.Model(&models.Attendee{}).
		Where("event_id = ? AND status = ?", eventID, models.AttendeeActive).
		Count(&active)
	stats["active"] = active

	// Cancelled
	var cancelled int64
	a.db.Model(&models.Attendee{}).
		Where("event_id = ? AND status = ?", eventID, models.AttendeeCancelled).
		Count(&cancelled)
	stats["cancelled"] = cancelled

	return stats, nil
}

func (a *attendeeRepoPG) GetHostAttendeeStats(hostID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total attendees across all host events
	var total int64
	a.db.Model(&models.Attendee{}).
		Joins("JOIN events ON attendees.event_id = events.id").
		Where("events.host_id = ?", hostID).
		Count(&total)
	stats["total"] = total

	// Checked in
	var checkedIn int64
	a.db.Model(&models.Attendee{}).
		Joins("JOIN events ON attendees.event_id = events.id").
		Where("events.host_id = ? AND attendees.status = ?", hostID, models.AttendeeCheckedIn).
		Count(&checkedIn)
	stats["checked_in"] = checkedIn

	// Active
	var active int64
	a.db.Model(&models.Attendee{}).
		Joins("JOIN events ON attendees.event_id = events.id").
		Where("events.host_id = ? AND attendees.status = ?", hostID, models.AttendeeActive).
		Count(&active)
	stats["active"] = active

	// Cancelled
	var cancelled int64
	a.db.Model(&models.Attendee{}).
		Joins("JOIN events ON attendees.event_id = events.id").
		Where("events.host_id = ? AND attendees.status = ?", hostID, models.AttendeeCancelled).
		Count(&cancelled)
	stats["cancelled"] = cancelled

	return stats, nil
}

func (a *attendeeRepoPG) GetByHostIDWithFilters(hostID uuid.UUID, limit, offset int, eventID *uuid.UUID, ticketType, status, search string) ([]models.Attendee, error) {
	var attendees []models.Attendee
	query := a.db.Preload("User").Preload("Event").Preload("Ticket").Preload("Ticket.TicketType").
		Joins("JOIN events ON attendees.event_id = events.id").
		Joins("JOIN tickets ON attendees.ticket_id = tickets.id").
		Joins("JOIN ticket_types ON tickets.ticket_type_id = ticket_types.id").
		Where("events.host_id = ?", hostID)

	// Apply filters
	if eventID != nil {
		query = query.Where("attendees.event_id = ?", *eventID)
	}

	if ticketType != "" {
		query = query.Where("ticket_types.name = ?", ticketType)
	}

	if status != "" {
		query = query.Where("attendees.status = ?", status)
	}

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ? OR LOWER(events.title) LIKE ?",
			searchPattern, searchPattern, searchPattern,
		).Joins("JOIN users ON attendees.user_id = users.id")
	} else {
		query = query.Joins("JOIN users ON attendees.user_id = users.id")
	}

	err := query.Order("attendees.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&attendees).Error
	return attendees, err
}

func (a *attendeeRepoPG) GetHostAttendeesTotalCountWithFilters(hostID uuid.UUID, eventID *uuid.UUID, ticketType, status, search string) (int64, error) {
	var count int64
	query := a.db.Model(&models.Attendee{}).
		Joins("JOIN events ON attendees.event_id = events.id").
		Joins("JOIN tickets ON attendees.ticket_id = tickets.id").
		Joins("JOIN ticket_types ON tickets.ticket_type_id = ticket_types.id").
		Where("events.host_id = ?", hostID)

	// Apply filters
	if eventID != nil {
		query = query.Where("attendees.event_id = ?", *eventID)
	}

	if ticketType != "" {
		query = query.Where("ticket_types.name = ?", ticketType)
	}

	if status != "" {
		query = query.Where("attendees.status = ?", status)
	}

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ? OR LOWER(events.title) LIKE ?",
			searchPattern, searchPattern, searchPattern,
		).Joins("JOIN users ON attendees.user_id = users.id")
	} else {
		query = query.Joins("JOIN users ON attendees.user_id = users.id")
	}

	err := query.Count(&count).Error
	return count, err
}
