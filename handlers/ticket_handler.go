
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
