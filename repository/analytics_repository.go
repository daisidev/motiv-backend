package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"gorm.io/gorm"
)

type AnalyticsRepository interface {
	// Event Views
	RecordEventView(view *models.EventView) error
	GetEventViews(eventID uuid.UUID, startDate, endDate time.Time) ([]models.EventView, error)

	// Event Analytics
	GetEventAnalytics(eventID uuid.UUID) (*models.EventAnalytics, error)
	UpdateEventAnalytics(analytics *models.EventAnalytics) error

	// Host Analytics
	GetHostAnalytics(hostID uuid.UUID) (*models.HostAnalytics, error)
	UpdateHostAnalytics(analytics *models.HostAnalytics) error

	// Dashboard Stats
	GetHostDashboardStats(hostID uuid.UUID) (map[string]interface{}, error)
	GetEventPerformanceStats(eventID uuid.UUID) (map[string]interface{}, error)
	GetMonthlyRevenueStats(hostID uuid.UUID, year int) ([]map[string]interface{}, error)
}

type analyticsRepoPG struct {
	db *gorm.DB
}

func NewAnalyticsRepoPG(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepoPG{db: db}
}

func (a *analyticsRepoPG) RecordEventView(view *models.EventView) error {
	return a.db.Create(view).Error
}

func (a *analyticsRepoPG) GetEventViews(eventID uuid.UUID, startDate, endDate time.Time) ([]models.EventView, error) {
	var views []models.EventView
	err := a.db.Where("event_id = ? AND viewed_at BETWEEN ? AND ?", eventID, startDate, endDate).
		Find(&views).Error
	return views, err
}

func (a *analyticsRepoPG) GetEventAnalytics(eventID uuid.UUID) (*models.EventAnalytics, error) {
	var analytics models.EventAnalytics
	err := a.db.Where("event_id = ?", eventID).First(&analytics).Error
	if err == gorm.ErrRecordNotFound {
		// Create new analytics record
		analytics = models.EventAnalytics{
			EventID: eventID,
		}
		err = a.db.Create(&analytics).Error
	}
	return &analytics, err
}

func (a *analyticsRepoPG) UpdateEventAnalytics(analytics *models.EventAnalytics) error {
	analytics.LastUpdated = time.Now()
	return a.db.Save(analytics).Error
}

func (a *analyticsRepoPG) GetHostAnalytics(hostID uuid.UUID) (*models.HostAnalytics, error) {
	var analytics models.HostAnalytics
	err := a.db.Where("host_id = ?", hostID).First(&analytics).Error
	if err == gorm.ErrRecordNotFound {
		// Create new analytics record
		analytics = models.HostAnalytics{
			HostID: hostID,
		}
		err = a.db.Create(&analytics).Error
	}
	return &analytics, err
}

func (a *analyticsRepoPG) UpdateHostAnalytics(analytics *models.HostAnalytics) error {
	analytics.LastUpdated = time.Now()
	return a.db.Save(analytics).Error
}

func (a *analyticsRepoPG) GetHostDashboardStats(hostID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total events
	var totalEvents int64
	a.db.Model(&models.Event{}).Where("host_id = ?", hostID).Count(&totalEvents)
	stats["total_events"] = totalEvents

	// Total revenue
	var totalRevenue float64
	a.db.Model(&models.Payment{}).
		Where("event_id IN (SELECT id FROM events WHERE host_id = ?) AND status = ?", hostID, models.PaymentCompleted).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRevenue)
	stats["total_revenue"] = totalRevenue

	// Total attendees
	var totalAttendees int64
	a.db.Model(&models.Attendee{}).
		Joins("JOIN events ON attendees.event_id = events.id").
		Where("events.host_id = ?", hostID).
		Count(&totalAttendees)
	stats["total_attendees"] = totalAttendees

	// This month's revenue
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var monthlyRevenue float64
	a.db.Model(&models.Payment{}).
		Where("event_id IN (SELECT id FROM events WHERE host_id = ?) AND status = ? AND created_at >= ?",
			hostID, models.PaymentCompleted, startOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&monthlyRevenue)
	stats["monthly_revenue"] = monthlyRevenue

	return stats, nil
}

func (a *analyticsRepoPG) GetEventPerformanceStats(eventID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total views
	var totalViews int64
	a.db.Model(&models.EventView{}).Where("event_id = ?", eventID).Count(&totalViews)
	stats["total_views"] = totalViews

	// Unique views
	var uniqueViews int64
	a.db.Model(&models.EventView{}).
		Where("event_id = ?", eventID).
		Distinct("COALESCE(user_id, ip_address)").
		Count(&uniqueViews)
	stats["unique_views"] = uniqueViews

	// Tickets sold
	var ticketsSold int64
	a.db.Model(&models.Ticket{}).Where("event_id = ?", eventID).Count(&ticketsSold)
	stats["tickets_sold"] = ticketsSold

	// Revenue
	var revenue float64
	a.db.Model(&models.Payment{}).
		Where("event_id = ? AND status = ?", eventID, models.PaymentCompleted).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&revenue)
	stats["revenue"] = revenue

	// Conversion rate
	var conversionRate float64
	if uniqueViews > 0 {
		conversionRate = float64(ticketsSold) / float64(uniqueViews) * 100
	}
	stats["conversion_rate"] = conversionRate

	return stats, nil
}

func (a *analyticsRepoPG) GetMonthlyRevenueStats(hostID uuid.UUID, year int) ([]map[string]interface{}, error) {
	var results []struct {
		Month   int
		Revenue float64
		Count   int64
	}

	err := a.db.Model(&models.Payment{}).
		Select("EXTRACT(MONTH FROM created_at) as month, COALESCE(SUM(amount), 0) as revenue, COUNT(*) as count").
		Where("event_id IN (SELECT id FROM events WHERE host_id = ?) AND status = ? AND EXTRACT(YEAR FROM created_at) = ?",
			hostID, models.PaymentCompleted, year).
		Group("EXTRACT(MONTH FROM created_at)").
		Order("month").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// Convert to map format
	var monthlyStats []map[string]interface{}
	for _, result := range results {
		monthlyStats = append(monthlyStats, map[string]interface{}{
			"month":   result.Month,
			"revenue": result.Revenue,
			"count":   result.Count,
		})
	}

	return monthlyStats, nil
}
