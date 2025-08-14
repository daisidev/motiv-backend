package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
)

// TicketHandler handles ticket-related requests
type TicketHandler struct {
	ticketService services.TicketService
}

func NewTicketHandler(ticketService services.TicketService) *TicketHandler {
	return &TicketHandler{ticketService}
}

// PurchaseTicket handles purchasing a ticket
func (h *TicketHandler) PurchaseTicket(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	var request struct {
		TicketTypeID string `json:"ticketTypeId"`
		Quantity     int    `json:"quantity"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ticketTypeID, err := uuid.Parse(request.TicketTypeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ticket type ID"})
	}

	// In a real application, you would have more logic here, e.g., payment processing, checking ticket availability, etc.

	// For simplicity, we'll just create a ticket directly
	ticket := &models.Ticket{
		UserID:       userID,
		TicketTypeID: ticketTypeID,
	}

	if err := h.ticketService.PurchaseTicket(ticket); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to purchase ticket"})
	}

	return c.Status(fiber.StatusCreated).JSON(ticket)
}

// RSVPFreeEvent handles RSVP for free events
func (h *TicketHandler) RSVPFreeEvent(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	var request struct {
		EventID          string `json:"eventId"`
		AttendeeFullName string `json:"attendeeFullName"`
		AttendeeEmail    string `json:"attendeeEmail"`
		AttendeePhone    string `json:"attendeePhone"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	eventID, err := uuid.Parse(request.EventID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	// Get the free ticket type for this event
	ticketTypes, err := h.ticketService.GetTicketTypesByEventID(eventID)
	if err != nil || len(ticketTypes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No ticket types found for this event"})
	}

	// Find the free ticket type (price = 0)
	var freeTicketType *models.TicketType
	for _, tt := range ticketTypes {
		if tt.Price == 0 {
			freeTicketType = tt
			break
		}
	}

	if freeTicketType == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No free ticket type found for this event"})
	}

	// Check if user already has a ticket for this event
	existingTickets, err := h.ticketService.GetTicketsByUserID(userID)
	if err == nil {
		for _, ticket := range existingTickets {
			if ticket.EventID == eventID {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "You already have a ticket for this event"})
			}
		}
	}

	// Create the free ticket
	ticket := &models.Ticket{
		EventID:          eventID,
		UserID:           userID,
		TicketTypeID:     freeTicketType.ID,
		PaymentReference: "FREE_RSVP",
		AttendeeFullName: request.AttendeeFullName,
		AttendeeEmail:    request.AttendeeEmail,
		AttendeePhone:    request.AttendeePhone,
		Quantity:         1,
	}

	if err := h.ticketService.CreateTicketWithQR(ticket); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create RSVP"})
	}

	// Update sold quantity for the ticket type
	if err := h.ticketService.UpdateSoldQuantity(freeTicketType.ID, 1); err != nil {
		// Log the error but don't fail the request
		// log.Printf("Failed to update sold quantity: %v", err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "RSVP successful",
		"ticket":  ticket,
	})
}
