package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hidenkeys/motiv-backend/services"
	"github.com/google/uuid"
)

type AnalyticsHandler struct {
	analyticsService services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GET /api/v1/hosts/me/analytics/dashboard
func (h *AnalyticsHandler) GetHostDashboard(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse user ID",
		})
	}
	
	stats, err := h.analyticsService.GetHostDashboardStats(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get dashboard stats",
		})
	}
	
	return c.JSON(fiber.Map{
		"data": stats,
	})
}

// GET /api/v1/hosts/me/analytics/revenue?year=2024
func (h *AnalyticsHandler) GetMonthlyRevenue(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse user ID",
		})
	}
	
	yearStr := c.Query("year", strconv.Itoa(time.Now().Year()))
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid year parameter",
		})
	}
	
	stats, err := h.analyticsService.GetMonthlyRevenueStats(userID, year)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get revenue stats",
		})
	}
	
	return c.JSON(fiber.Map{
		"data": stats,
	})
}

// GET /api/v1/events/:id/analytics
func (h *AnalyticsHandler) GetEventAnalytics(c *fiber.Ctx) error {
	eventIDStr := c.Params("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid event ID",
		})
	}
	
	analytics, err := h.analyticsService.GetEventAnalytics(eventID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get event analytics",
		})
	}
	
	performanceStats, err := h.analyticsService.GetEventPerformanceStats(eventID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get performance stats",
		})
	}
	
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"analytics":    analytics,
			"performance":  performanceStats,
		},
	})
}

// POST /api/v1/events/:id/view
func (h *AnalyticsHandler) RecordEventView(c *fiber.Ctx) error {
	eventIDStr := c.Params("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid event ID",
		})
	}
	
	// Get user ID if authenticated (optional)
	var userID *uuid.UUID
	if userValue := c.Locals("user"); userValue != nil {
		if user, ok := userValue.(*jwt.Token); ok {
			claims := user.Claims.(jwt.MapClaims)
			if userIDStr, ok := claims["user_id"].(string); ok {
				if uid, err := uuid.Parse(userIDStr); err == nil {
					userID = &uid
				}
			}
		}
	}
	
	// Get IP and User Agent
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	
	err = h.analyticsService.RecordEventView(eventID, userID, ipAddress, userAgent)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to record view",
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "View recorded successfully",
	})
}