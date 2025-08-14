package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/services"
)

type AttendeeHandler struct {
	attendeeService services.AttendeeService
	eventService    services.EventService
}

func NewAttendeeHandler(attendeeService services.AttendeeService, eventService services.EventService) *AttendeeHandler {
	return &AttendeeHandler{
		attendeeService: attendeeService,
		eventService:    eventService,
	}
}

// GetEventAttendees handles retrieving attendees for a specific event
func (h *AttendeeHandler) GetEventAttendees(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	// Verify the host owns this event
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	hostID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	event, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	if event.HostID != hostID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to view attendees for this event"})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	attendees, err := h.attendeeService.GetEventAttendees(eventID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get attendees"})
	}

	// Get total count for pagination
	totalCount, err := h.attendeeService.GetEventAttendeesTotalCount(eventID)
	if err != nil {
		log.Printf("Failed to get total attendee count: %v", err)
		totalCount = int64(len(attendees))
	}

	// Get stats
	stats, err := h.attendeeService.GetEventAttendeeStats(eventID)
	if err != nil {
		log.Printf("Failed to get attendee stats: %v", err)
		stats = make(map[string]int64)
	}

	return c.JSON(fiber.Map{
		"data":    attendees,
		"total":   totalCount,
		"page":    page,
		"limit":   limit,
		"hasMore": int64(offset+limit) < totalCount,
		"stats":   stats,
	})
}

// CheckInAttendee handles checking in an attendee via QR code
func (h *AttendeeHandler) CheckInAttendee(c *fiber.Ctx) error {
	var req struct {
		QRCode  string `json:"qrCode" validate:"required"`
		EventID string `json:"eventId" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get host ID from JWT
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	hostID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	// Verify the host owns this event
	event, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	if event.HostID != hostID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to check in attendees for this event"})
	}

	// Find attendee by QR code (assuming QR code is the ticket ID or a unique identifier)
	result, err := h.attendeeService.CheckInByQRCode(req.QRCode, eventID, hostID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

// GetHostAttendees handles retrieving all attendees for a host's events
func (h *AttendeeHandler) GetHostAttendees(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	hostID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	attendees, err := h.attendeeService.GetHostAttendees(hostID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get attendees"})
	}

	// Get total count for pagination
	totalCount, err := h.attendeeService.GetHostAttendeesTotalCount(hostID)
	if err != nil {
		log.Printf("Failed to get total attendee count: %v", err)
		totalCount = int64(len(attendees))
	}

	// Get stats
	stats, err := h.attendeeService.GetHostAttendeeStats(hostID)
	if err != nil {
		log.Printf("Failed to get host attendee stats: %v", err)
		stats = make(map[string]int64)
	}

	return c.JSON(fiber.Map{
		"data":    attendees,
		"total":   totalCount,
		"page":    page,
		"limit":   limit,
		"hasMore": int64(offset+limit) < totalCount,
		"stats":   stats,
	})
}
