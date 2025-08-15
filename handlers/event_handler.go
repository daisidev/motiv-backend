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
	eventService  services.EventService
	ticketService services.TicketService
}

func NewEventHandler(eventService services.EventService, ticketService services.TicketService) *EventHandler {
	return &EventHandler{eventService, ticketService}
}

// GetAllEvents handles retrieving all events with pagination
func (h *EventHandler) GetAllEvents(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "12"))
	search := c.Query("search", "")
	tags := c.Query("tags", "")
	location := c.Query("location", "")
	eventType := c.Query("event_type", "") // "free" or "ticketed"
	dateFrom := c.Query("date_from", "")
	dateTo := c.Query("date_to", "")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 12
	}

	params := services.EventQueryParams{
		Page:      page,
		Limit:     limit,
		Search:    search,
		Tags:      tags,
		Location:  location,
		EventType: eventType,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
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

	// Validate event type and related ticket type constraints using tagged switch
	switch req.EventType {
	case "ticketed":
		if len(req.TicketTypes) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ticketed events must have at least one ticket type"})
		}
	case "free":
		if len(req.TicketTypes) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Free events cannot have ticket types"})
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Event type must be 'ticketed' or 'free'"})
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start date format. Use YYYY-MM-DD"})
	}

	// Determine status - default to active if not specified
	status := models.ActiveEvent
	if req.Status != "" {
		switch req.Status {
		case "draft":
			status = models.DraftEvent
		case "active":
			status = models.ActiveEvent
		case "cancelled":
			status = models.CancelledEvent
		default:
			status = models.ActiveEvent
		}
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
		Status:         status,
	}

	// Add location data if provided
	if req.LocationData != nil {
		newEvent.Latitude = &req.LocationData.Coordinates.Lat
		newEvent.Longitude = &req.LocationData.Coordinates.Lng
		newEvent.ManualDescription = req.LocationData.ManualDescription
		if req.LocationData.PlaceID != nil {
			newEvent.PlaceID = req.LocationData.PlaceID
		}
	}

	// Create ticket types slice for possible ticketed event
	var ticketTypes []models.TicketType
	if req.EventType == "ticketed" {
		for _, ticketReq := range req.TicketTypes {
			ticketType := models.TicketType{
				Name:          ticketReq.Name,
				Price:         ticketReq.Price,
				Description:   ticketReq.Description,
				TotalQuantity: ticketReq.TotalQuantity,
				SoldQuantity:  0,
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

	// Create ticket types or default free ticket type depending on event type
	switch req.EventType {
	case "ticketed":
		for i := range ticketTypes {
			ticketTypes[i].EventID = newEvent.ID
			err = h.ticketService.CreateTicketType(&ticketTypes[i])
			if err != nil {
				log.Printf("Error creating ticket type: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create ticket types"})
			}
		}
		// Attach ticket types to the event for response
		newEvent.TicketTypes = ticketTypes
	case "free":
		// Create a default free ticket type for free events
		freeTicketType := models.TicketType{
			EventID:       newEvent.ID,
			Name:          "Free Entry",
			Price:         0,
			Description:   "Free admission to this event",
			TotalQuantity: 1000, // Default capacity for free events
			SoldQuantity:  0,
		}

		err = h.ticketService.CreateTicketType(&freeTicketType)
		if err != nil {
			log.Printf("Error creating free ticket type: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create free ticket type"})
		}

		// Attach the free ticket type to the event for response
		newEvent.TicketTypes = []models.TicketType{freeTicketType}
	}

	return c.Status(fiber.StatusCreated).JSON(newEvent)
}

// UpdateEvent handles updating an event
func (h *EventHandler) UpdateEvent(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	var req models.CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request: " + err.Error()})
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

	// Parse start date if provided
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start date format. Use YYYY-MM-DD"})
		}
		event.StartDate = startDate
	}

	// Update fields
	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if req.StartTime != "" {
		event.StartTime = req.StartTime
	}
	if req.EndTime != "" {
		event.EndTime = req.EndTime
	}
	if req.Location != "" {
		event.Location = req.Location
	}
	
	// Update location data if provided
	if req.LocationData != nil {
		event.Latitude = &req.LocationData.Coordinates.Lat
		event.Longitude = &req.LocationData.Coordinates.Lng
		event.ManualDescription = req.LocationData.ManualDescription
		if req.LocationData.PlaceID != nil {
			event.PlaceID = req.LocationData.PlaceID
		}
	}
	
	if req.Tags != nil {
		event.Tags = req.Tags
	}
	if req.BannerImageURL != "" {
		event.BannerImageURL = req.BannerImageURL
	}
	if req.EventType != "" {
		event.EventType = req.EventType
	}

	// Handle status update - default to active if not specified in request
	if req.Status != "" {
		switch req.Status {
		case "draft":
			event.Status = models.DraftEvent
		case "active":
			event.Status = models.ActiveEvent
		case "cancelled":
			event.Status = models.CancelledEvent
		default:
			event.Status = models.ActiveEvent
		}
	} else {
		// Default to active for updates if no status is specified
		event.Status = models.ActiveEvent
	}

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

// GetSearchSuggestions handles getting search suggestions
func (h *EventHandler) GetSearchSuggestions(c *fiber.Ctx) error {
	query := c.Query("q", "")

	if len(query) < 2 {
		return c.JSON([]string{})
	}

	suggestions, err := h.eventService.GetSearchSuggestions(query, 3)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get suggestions"})
	}

	return c.JSON(suggestions)
}
