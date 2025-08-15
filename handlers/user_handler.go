package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userService     services.UserService
	wishlistService services.WishlistService
	ticketService   services.TicketService
}

func NewUserHandler(userService services.UserService, wishlistService services.WishlistService, ticketService services.TicketService) *UserHandler {
	return &UserHandler{userService, wishlistService, ticketService}
}

// GetMe handles retrieving the current user's profile
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	currentUser, err := h.userService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(currentUser)
}

// UpdateMe handles updating the current user's profile
func (h *UserHandler) UpdateMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	var updateUser models.User
	if err := c.BodyParser(&updateUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	currentUser, err := h.userService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	currentUser.Name = updateUser.Name
	currentUser.Avatar = updateUser.Avatar

	if err := h.userService.UpdateUser(currentUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user"})
	}

	return c.JSON(currentUser)
}

// GetMyTickets handles retrieving the current user's tickets
func (h *UserHandler) GetMyTickets(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	// Validate that the user exists
	_, err = h.userService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	tickets, err := h.ticketService.GetTicketsByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get tickets"})
	}

	// Additional logging for debugging
	log.Printf("Returning %d tickets for user %s", len(tickets), userID.String())
	for i, ticket := range tickets {
		log.Printf("Ticket %d: ID=%s, EventID=%s, EventTitle=%s", 
			i, ticket.ID.String(), ticket.EventID.String(), 
			func() string {
				if ticket.Event.Title != "" {
					return ticket.Event.Title
				}
				return "EMPTY_TITLE"
			}())
	}

	return c.JSON(tickets)
}

// GetMyWishlist handles retrieving the current user's wishlist
func (h *UserHandler) GetMyWishlist(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	wishlistItems, err := h.wishlistService.GetWishlistItemsByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get wishlist"})
	}

	return c.JSON(wishlistItems)
}

// AddToMyWishlist handles adding an event to the current user's wishlist
func (h *UserHandler) AddToMyWishlist(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	var request struct {
		EventID string `json:"event_id"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Debug logging
	fmt.Printf("AddToMyWishlist - Received EventID: '%s'\n", request.EventID)
	fmt.Printf("AddToMyWishlist - EventID length: %d\n", len(request.EventID))

	eventID, err := uuid.Parse(request.EventID)
	if err != nil {
		fmt.Printf("AddToMyWishlist - UUID Parse Error: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	wishlist := &models.Wishlist{
		UserID:  userID,
		EventID: eventID,
	}

	if err := h.wishlistService.AddToWishlist(wishlist); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add to wishlist"})
	}

	return c.SendStatus(fiber.StatusOK)
}

// RemoveFromMyWishlist handles removing an event from the current user's wishlist
func (h *UserHandler) RemoveFromMyWishlist(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	// Get event_id from query parameter
	eventIDStr := c.Query("event_id")
	if eventIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "event_id query parameter is required"})
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	if err := h.wishlistService.RemoveFromWishlist(userID, eventID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove from wishlist"})
	}

	return c.SendStatus(fiber.StatusOK)
}

// CheckWishlistStatus handles checking if an event is in the user's wishlist
func (h *UserHandler) CheckWishlistStatus(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	// Get event_id from query parameter
	eventIDStr := c.Query("event_id")
	if eventIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "event_id query parameter is required"})
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	isInWishlist, err := h.wishlistService.IsInWishlist(userID, eventID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check wishlist status"})
	}

	return c.JSON(fiber.Map{"is_in_wishlist": isInWishlist})
}

// GetMyTicket handles retrieving a specific ticket by ID for the current user
func (h *UserHandler) GetMyTicket(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	// Get ticket ID from URL parameter
	ticketIDStr := c.Params("id")
	if ticketIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ticket ID is required"})
	}

	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ticket ID"})
	}

	// Get the ticket and verify it belongs to the user
	ticket, err := h.ticketService.GetTicketByID(ticketID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Ticket not found"})
	}

	// Verify the ticket belongs to the current user
	if ticket.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	return c.JSON(ticket)
}

// GetMyTicketsDebug handles debugging ticket data issues
func (h *UserHandler) GetMyTicketsDebug(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	tickets, err := h.ticketService.GetTicketsByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get tickets"})
	}

	// Return detailed debug information
	debugInfo := make([]map[string]interface{}, len(tickets))
	for i, ticket := range tickets {
		debugInfo[i] = map[string]interface{}{
			"ticket_id":           ticket.ID.String(),
			"event_id":            ticket.EventID.String(),
			"event_loaded":        ticket.Event.ID != uuid.Nil,
			"event_title":         ticket.Event.Title,
			"event_location":      ticket.Event.Location,
			"event_start_date":    ticket.Event.StartDate,
			"ticket_type_loaded":  ticket.TicketType.ID != uuid.Nil,
			"ticket_type_name":    ticket.TicketType.Name,
			"attendee_name":       ticket.AttendeeFullName,
			"payment_reference":   ticket.PaymentReference,
		}
	}

	return c.JSON(fiber.Map{
		"total_tickets": len(tickets),
		"debug_info":    debugInfo,
	})
}
