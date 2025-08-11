package services

import (
	"time"

	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
	"github.com/google/uuid"
)

type AnalyticsService interface {
	// Event Views
	RecordEventView(eventID uuid.UUID, userID *uuid.UUID, ipAddress, userAgent string) error
	
	// Dashboard Stats
	GetHostDashboardStats(hostID uuid.UUID) (map[string]interface{}, error)
	GetEventAnalytics(eventID uuid.UUID) (*models.EventAnalytics, error)
	GetHostAnalytics(hostID uuid.UUID) (*models.HostAnalytics, error)
	
	// Performance Stats
	GetMonthlyRevenueStats(hostID uuid.UUID, year int) ([]map[string]interface{}, error)
	GetEventPerformanceStats(eventID uuid.UUID) (map[string]interface{}, error)
	
	// Update Analytics
	UpdateEventAnalytics(eventID uuid.UUID) error
	UpdateHostAnalytics(hostID uuid.UUID) error
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	paymentRepo   repository.PaymentRepository
	attendeeRepo  repository.AttendeeRepository
	reviewRepo    repository.ReviewRepository
}

func NewAnalyticsService(
	analyticsRepo repository.AnalyticsRepository,
	paymentRepo repository.PaymentRepository,
	attendeeRepo repository.AttendeeRepository,
	reviewRepo repository.ReviewRepository,
) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		paymentRepo:   paymentRepo,
		attendeeRepo:  attendeeRepo,
		reviewRepo:    reviewRepo,
	}
}

func (s *analyticsService) RecordEventView(eventID uuid.UUID, userID *uuid.UUID, ipAddress, userAgent string) error {
	view := &models.EventView{
		EventID:   eventID,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ViewedAt:  time.Now(),
	}
	
	return s.analyticsRepo.RecordEventView(view)
}

func (s *analyticsService) GetHostDashboardStats(hostID uuid.UUID) (map[string]interface{}, error) {
	return s.analyticsRepo.GetHostDashboardStats(hostID)
}

func (s *analyticsService) GetEventAnalytics(eventID uuid.UUID) (*models.EventAnalytics, error) {
	return s.analyticsRepo.GetEventAnalytics(eventID)
}

func (s *analyticsService) GetHostAnalytics(hostID uuid.UUID) (*models.HostAnalytics, error) {
	return s.analyticsRepo.GetHostAnalytics(hostID)
}

func (s *analyticsService) GetMonthlyRevenueStats(hostID uuid.UUID, year int) ([]map[string]interface{}, error) {
	return s.analyticsRepo.GetMonthlyRevenueStats(hostID, year)
}

func (s *analyticsService) GetEventPerformanceStats(eventID uuid.UUID) (map[string]interface{}, error) {
	return s.analyticsRepo.GetEventPerformanceStats(eventID)
}

func (s *analyticsService) UpdateEventAnalytics(eventID uuid.UUID) error {
	analytics, err := s.analyticsRepo.GetEventAnalytics(eventID)
	if err != nil {
		return err
	}
	
	// Get performance stats
	stats, err := s.analyticsRepo.GetEventPerformanceStats(eventID)
	if err != nil {
		return err
	}
	
	// Update analytics
	analytics.TotalViews = int(stats["total_views"].(int64))
	analytics.UniqueViews = int(stats["unique_views"].(int64))
	analytics.TicketsSold = int(stats["tickets_sold"].(int64))
	analytics.Revenue = stats["revenue"].(float64)
	analytics.ConversionRate = stats["conversion_rate"].(float64)
	
	return s.analyticsRepo.UpdateEventAnalytics(analytics)
}

func (s *analyticsService) UpdateHostAnalytics(hostID uuid.UUID) error {
	analytics, err := s.analyticsRepo.GetHostAnalytics(hostID)
	if err != nil {
		return err
	}
	
	// Get dashboard stats
	stats, err := s.analyticsRepo.GetHostDashboardStats(hostID)
	if err != nil {
		return err
	}
	
	// Get host rating stats
	_, averageRating, err := s.reviewRepo.GetHostRatingStats(hostID)
	if err != nil {
		return err
	}
	
	// Update analytics
	analytics.TotalEvents = int(stats["total_events"].(int64))
	analytics.TotalRevenue = stats["total_revenue"].(float64)
	analytics.TotalAttendees = int(stats["total_attendees"].(int64))
	analytics.AverageRating = averageRating
	
	return s.analyticsRepo.UpdateHostAnalytics(analytics)
}