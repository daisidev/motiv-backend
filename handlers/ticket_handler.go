package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
)

// TicketHandler handles ticket-related requests
type TicketHandler struct {
	ticketService services.TicketService
	eventService  services.EventService
	userService   services.UserService
	emailService  services.EmailService
}

func NewTicketHandler(ticketService services.TicketService, eventService services.EventService, userService services.UserService, emailService services.EmailService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
		eventService:  eventService,
		userService:   userService,
		emailService:  emailService,
	}
}

// PurchaseTicket handles purchasing a ticket
// NOTE: This should NOT create tickets directly. Tickets should only be created via webhook confirmation.
func (h *TicketHandler) PurchaseTicket(c *fiber.Ctx) error {
	log.Printf("🚨 SECURITY ALERT: /tickets/purchase route called - THIS SHOULD NOT HAPPEN FOR PAID EVENTS!")
	log.Printf("🚨 CLIENT INFO: IP=%s, User-Agent=%s", c.IP(), c.Get("User-Agent"))

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		log.Printf("❌ TICKET PURCHASE ERROR: Failed to parse user ID: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	var request struct {
		TicketTypeID string `json:"ticketTypeId"`
		Quantity     int    `json:"quantity"`
	}

	if err := c.BodyParser(&request); err != nil {
		log.Printf("❌ TICKET PURCHASE ERROR: Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	log.Printf("🚫 BLOCKING DIRECT PURCHASE: User %s attempted to purchase tickets directly without payment", userID.String())
	log.Printf("🚫 REQUEST DETAILS: TicketTypeID=%s, Quantity=%d", request.TicketTypeID, request.Quantity)
	log.Printf("🚫 THIS INDICATES: Frontend is trying to bypass payment verification!")

	// IMPORTANT: Do not create tickets directly! This bypasses payment verification.
	// Instead, return an error directing them to use the proper payment flow.
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error":    "Direct ticket purchase is not allowed. Please use the payment initiation flow at /api/v1/payments/initiate",
		"message":  "Tickets can only be created after successful payment confirmation via webhook",
		"redirect": "/api/v1/payments/initiate",
	})

	// OLD CODE REMOVED - This was creating tickets without payment verification:
	// ticketTypeID, err := uuid.Parse(request.TicketTypeID)
	// if err != nil {
	//     return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ticket type ID"})
	// }
	//
	// ticket := &models.Ticket{
	//     UserID:       userID,
	//     TicketTypeID: ticketTypeID,
	// }
	//
	// if err := h.ticketService.PurchaseTicket(ticket); err != nil {
	//     return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to purchase ticket"})
	// }
	//
	// return c.Status(fiber.StatusCreated).JSON(ticket)
}

// RSVPFreeEvent handles RSVP for free events
func (h *TicketHandler) RSVPFreeEvent(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		log.Printf("❌ FREE RSVP ERROR: Failed to parse user ID: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	log.Printf("🆓 FREE RSVP: Starting free event RSVP for user: %s", userID.String())

	var request struct {
		EventID          string `json:"eventId"`
		AttendeeFullName string `json:"attendeeFullName"`
		AttendeeEmail    string `json:"attendeeEmail"`
		AttendeePhone    string `json:"attendeePhone"`
	}

	if err := c.BodyParser(&request); err != nil {
		log.Printf("❌ FREE RSVP ERROR: Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	log.Printf("🆓 RSVP DETAILS: EventID=%s, Attendee=%s (%s)", request.EventID, request.AttendeeFullName, request.AttendeeEmail)

	eventID, err := uuid.Parse(request.EventID)
	if err != nil {
		log.Printf("❌ FREE RSVP ERROR: Invalid event ID format: %s", request.EventID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	// Get the free ticket type for this event
	ticketTypes, err := h.ticketService.GetTicketTypesByEventID(eventID)
	if err != nil || len(ticketTypes) == 0 {
		log.Printf("🆓 NO TICKET TYPES: No ticket types found for event %s, checking if it's a free event", eventID.String())

		// For free events, create a default free ticket type if none exists
		// First check if this is a free event
		event, eventErr := h.eventService.GetEventByID(eventID)
		if eventErr != nil {
			log.Printf("❌ FREE RSVP ERROR: Event %s not found: %v", eventID.String(), eventErr)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Event not found"})
		}

		log.Printf("📅 EVENT VERIFIED: Event '%s' type: %s", event.Title, event.EventType)

		if event.EventType == "free" {
			log.Printf("🆓 CREATING FREE TICKET TYPE: Creating default free ticket type for event %s", eventID.String())

			// Create a default free ticket type
			freeTicketType := models.TicketType{
				EventID:       eventID,
				Name:          "Free Entry",
				Price:         0,
				Description:   "Free admission to this event",
				TotalQuantity: 1000, // Default capacity for free events
				SoldQuantity:  0,
			}

			createErr := h.ticketService.CreateTicketType(&freeTicketType)
			if createErr != nil {
				log.Printf("❌ FREE RSVP ERROR: Failed to create free ticket type: %v", createErr)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create free ticket type"})
			}

			log.Printf("✅ FREE TICKET TYPE CREATED: Default free ticket type created for event %s", eventID.String())
			ticketTypes = []*models.TicketType{&freeTicketType}
		} else {
			log.Printf("❌ FREE RSVP ERROR: Event %s is not a free event (type: %s)", eventID.String(), event.EventType)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No ticket types found for this event"})
		}
	}

	// Find the free ticket type (price = 0)
	var freeTicketType *models.TicketType
	for _, tt := range ticketTypes {
		if tt.Price == 0 {
			freeTicketType = tt
			log.Printf("🆓 FREE TICKET TYPE FOUND: %s (ID: %s, Available: %d)", tt.Name, tt.ID.String(), tt.TotalQuantity-tt.SoldQuantity)
			break
		}
	}

	if freeTicketType == nil {
		log.Printf("❌ FREE RSVP ERROR: No free ticket type found for event %s", eventID.String())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No free ticket type found for this event"})
	}

	// Check if user already has a ticket for this event
	existingTickets, err := h.ticketService.GetTicketsByUserID(userID)
	if err == nil {
		for _, ticket := range existingTickets {
			if ticket.EventID == eventID {
				log.Printf("❌ FREE RSVP ERROR: User %s already has a ticket for event %s", userID.String(), eventID.String())
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "You already have a ticket for this event"})
			}
		}
	}
	log.Printf("✅ USER VERIFICATION: User %s doesn't have existing ticket for event %s", userID.String(), eventID.String())

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

	log.Printf("🎫 CREATING FREE TICKET: Creating ticket for attendee %s", request.AttendeeFullName)

	if err := h.ticketService.CreateTicketWithQR(ticket); err != nil {
		log.Printf("❌ FREE RSVP ERROR: Failed to create RSVP ticket: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create RSVP"})
	}

	log.Printf("✅ FREE TICKET CREATED: Successfully created free ticket %s for user %s", ticket.ID.String(), userID.String())

	// Update sold quantity for the ticket type
	if err := h.ticketService.UpdateSoldQuantity(freeTicketType.ID, 1); err != nil {
		log.Printf("⚠️ FREE RSVP WARNING: Failed to update sold quantity: %v", err)
		// Log the error but don't fail the request
	} else {
		log.Printf("📊 SALES UPDATED: Updated sold quantity for free ticket type %s", freeTicketType.ID.String())
	}

	// Get event details and host for email notifications
	eventDetails, err := h.eventService.GetEventByID(eventID)
	if err == nil {
		log.Printf("📧 EMAIL NOTIFICATIONS: Starting email notifications for free ticket %s", ticket.ID.String())

		// Get user details
		userDetails, userErr := h.userService.GetUserByID(userID)
		if userErr == nil {
			// Get host details
			host, hostErr := h.userService.GetUserByID(eventDetails.HostID)
			if hostErr == nil {
				// Send ticket confirmation email to customer
				if emailErr := h.emailService.SendTicketConfirmation(ticket, eventDetails, userDetails); emailErr != nil {
					log.Printf("❌ EMAIL ERROR: Failed to send ticket confirmation email for free ticket %s: %v", ticket.ID.String(), emailErr)
					// Log the error but don't fail the request
				} else {
					log.Printf("✅ EMAIL SENT: Ticket confirmation email sent for free ticket %s", ticket.ID.String())
				}

				// Send notification email to host
				if emailErr := h.emailService.SendHostNotification(ticket, eventDetails, userDetails, host); emailErr != nil {
					log.Printf("❌ EMAIL ERROR: Failed to send host notification email for free ticket %s: %v", ticket.ID.String(), emailErr)
					// Log the error but don't fail the request
				} else {
					log.Printf("✅ EMAIL SENT: Host notification email sent for free ticket %s", ticket.ID.String())
				}
			} else {
				log.Printf("⚠️ EMAIL WARNING: Failed to get host details for notifications: %v", hostErr)
			}
		} else {
			log.Printf("⚠️ EMAIL WARNING: Failed to get user details for notifications: %v", userErr)
		}
	} else {
		log.Printf("⚠️ EMAIL WARNING: Failed to get event details for notifications: %v", err)
	}

	log.Printf("🎉 FREE RSVP SUCCESS: RSVP completed successfully for user %s, event %s", userID.String(), eventID.String())
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "RSVP successful",
		"ticket":  ticket,
	})
}
