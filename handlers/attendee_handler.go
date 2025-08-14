package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"strings"

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

	// Parse filter parameters
	eventIDStr := c.Query("event_id")
	ticketType := c.Query("ticket_type")
	status := c.Query("status")
	search := c.Query("search")

	var eventID *uuid.UUID
	if eventIDStr != "" {
		if parsedEventID, err := uuid.Parse(eventIDStr); err == nil {
			eventID = &parsedEventID
		}
	}

	attendees, err := h.attendeeService.GetHostAttendeesWithFilters(hostID, limit, offset, eventID, ticketType, status, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get attendees"})
	}

	// Get total count for pagination
	totalCount, err := h.attendeeService.GetHostAttendeesTotalCountWithFilters(hostID, eventID, ticketType, status, search)
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

// ExportHostAttendees handles exporting attendees to CSV
func (h *AttendeeHandler) ExportHostAttendees(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	hostID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	// Parse filter parameters
	eventIDStr := c.Query("event_id")
	ticketType := c.Query("ticket_type")
	status := c.Query("status")
	search := c.Query("search")

	var eventID *uuid.UUID
	if eventIDStr != "" {
		if parsedEventID, err := uuid.Parse(eventIDStr); err == nil {
			eventID = &parsedEventID
		}
	}

	// Get all attendees (no pagination for export)
	attendees, err := h.attendeeService.GetHostAttendeesWithFilters(hostID, 10000, 0, eventID, ticketType, status, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get attendees for export"})
	}

	// Create CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	headers := []string{"Name", "Email", "Phone", "Event", "Ticket Type", "Purchase Date", "Amount", "Status", "Check-in Status", "Check-in Time"}
	if err := writer.Write(headers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write CSV headers"})
	}

	// Write data
	for _, attendee := range attendees {
		checkinStatus := "Not Checked In"
		checkinTime := ""
		if attendee.Status == "checked_in" {
			checkinStatus = "Checked In"
			if attendee.CheckInTime != nil {
				checkinTime = attendee.CheckInTime.Format("2006-01-02 15:04:05")
			}
		}

		row := []string{
			attendee.Name,
			attendee.Email,
			attendee.Phone,
			attendee.EventTitle,
			attendee.TicketType,
			attendee.PurchaseDate,
			fmt.Sprintf("%.2f", attendee.Amount),
			strings.Title(strings.ReplaceAll(attendee.Status, "_", " ")),
			checkinStatus,
			checkinTime,
		}
		if err := writer.Write(row); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write CSV row"})
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to finalize CSV"})
	}

	// Set headers for file download
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=attendees.csv")

	return c.Send(buf.Bytes())
}
