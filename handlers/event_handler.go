
package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
)

// EventHandler handles event-related requests
type EventHandler struct {
	eventService services.EventService
}

func NewEventHandler(eventService services.EventService) *EventHandler {
	return &EventHandler{eventService}
}

// GetAllEvents handles retrieving all events with pagination
func (h *EventHandler) GetAllEvents(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "12"))
	search := c.Query("search", "")
	tags := c.Query("tags", "")
	location := c.Query("location", "")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	params := services.EventQueryParams{
		Page:     page,
		Limit:    limit,
		Search:   search,
		Tags:     tags,
		Location: location,
	}

	result, err := h.eventService.GetAllEventsWithPagination(params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get events"})
	}

	return c.JSON(result)
}

// GetEventByID handles retrieving an event by its ID
func (h *EventHandler) GetEventByID(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	event, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	return c.JSON(event)
}

// GetMyEvents handles retrieving events for the current host
func (h *EventHandler) GetMyEvents(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	hostID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	events, err := h.eventService.GetEventsByHostID(hostID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get events"})
	}

	return c.JSON(events)
}

// CreateEvent handles creating a new event
func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	hostID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	var req models.CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
	}

	// Validate event type
	if req.EventType != "ticketed" && req.EventType != "free" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Event type must be 'ticketed' or 'free'"})
	}

	// For ticketed events, validate that ticket types are provided
	if req.EventType == "ticketed" && len(req.TicketTypes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ticketed events must have at least one ticket type"})
	}

	// For free events, ensure no ticket types are provided
	if req.EventType == "free" && len(req.TicketTypes) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Free events cannot have ticket types"})
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start date format. Use YYYY-MM-DD"})
	}

	// Create the event
	newEvent := models.Event{
		Title:          req.Title,
		Description:    req.Description,
		StartDate:      startDate,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Location:       req.Location,
		Tags:           req.Tags,
		BannerImageURL: req.BannerImageURL,
		EventType:      req.EventType,
		HostID:         hostID,
		Status:         models.DraftEvent,
	}

	// Create ticket types for ticketed events
	var ticketTypes []models.TicketType
	if req.EventType == "ticketed" {
		for _, ticketReq := range req.TicketTypes {
			ticketType := models.TicketType{
				Name:            ticketReq.Name,
				Price:           ticketReq.Price,
				Description:     ticketReq.Description,
				TotalQuantity:   ticketReq.TotalQuantity,
				SoldQuantity:    0,
			}
			ticketTypes = append(ticketTypes, ticketType)
		}
	}

	// Create the event first
	err = h.eventService.CreateEvent(&newEvent)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create event"})
	}

	// For now, we'll return the created event
	// TODO: Add ticket type creation logic when ticket service is integrated
	createdEvent := &newEvent

	return c.Status(fiber.StatusCreated).JSON(createdEvent)
}

// UpdateEvent handles updating an event
func (h *EventHandler) UpdateEvent(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	var updateEvent models.Event
	if err := c.BodyParser(&updateEvent); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	event, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	hostID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	if event.HostID != hostID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to update this event"})
	}

	event.Title = updateEvent.Title
	event.Description = updateEvent.Description
	event.StartDate = updateEvent.StartDate
	event.StartTime = updateEvent.StartTime
	event.EndTime = updateEvent.EndTime
	event.Location = updateEvent.Location
	event.Tags = updateEvent.Tags
	event.BannerImageURL = updateEvent.BannerImageURL
	event.Status = updateEvent.Status

	if err := h.eventService.UpdateEvent(event); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update event"})
	}

	return c.JSON(event)
}

// DeleteEvent handles deleting an event
func (h *EventHandler) DeleteEvent(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	event, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	hostID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	if event.HostID != hostID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to delete this event"})
	}

	if err := h.eventService.DeleteEvent(eventID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete event"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
