package handlers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
)

type PaymentHandler struct {
	paymentService services.PaymentService
	ticketService  services.TicketService
	eventService   services.EventService
	userService    services.UserService
	emailService   services.EmailService
}

func NewPaymentHandler(paymentService services.PaymentService, ticketService services.TicketService, eventService services.EventService, userService services.UserService, emailService services.EmailService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		ticketService:  ticketService,
		eventService:   eventService,
		userService:    userService,
		emailService:   emailService,
	}
}

// GET /api/v1/hosts/me/payments/earnings
func (h *PaymentHandler) GetHostEarnings(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse user ID",
		})
	}

	earnings, err := h.paymentService.GetHostEarnings(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get earnings",
		})
	}

	return c.JSON(fiber.Map{
		"data": earnings,
	})
}

// GET /api/v1/hosts/me/payments/payouts
func (h *PaymentHandler) GetHostPayouts(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse user ID",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	payouts, err := h.paymentService.GetHostPayouts(userID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get payouts",
		})
	}

	return c.JSON(fiber.Map{
		"data": payouts,
	})
}

// GET /api/v1/hosts/me/payments/pending
func (h *PaymentHandler) GetPendingPayouts(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse user ID",
		})
	}

	payouts, err := h.paymentService.GetPendingPayouts(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get pending payouts",
		})
	}

	return c.JSON(fiber.Map{
		"data": payouts,
	})
}

// GET /api/v1/events/:id/revenue
func (h *PaymentHandler) GetEventRevenue(c *fiber.Ctx) error {
	eventIDStr := c.Params("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid event ID",
		})
	}

	revenue, err := h.paymentService.GetEventRevenue(eventID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get event revenue",
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"revenue": revenue,
		},
	})
}

type UpdatePaymentStatusRequest struct {
	Reference     string `json:"reference" validate:"required"`
	Status        string `json:"status" validate:"required"`
	FailureReason string `json:"failureReason,omitempty"`
}

// POST /api/v1/payments/initiate
func (h *PaymentHandler) InitiatePayment(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	// Validate that the user exists
	_, err = h.userService.GetUserByID(userID)
	if err != nil {
		log.Printf("User with ID %s not found: %v", userID, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	var req models.PaymentInitiationRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing payment initiation request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	// Verify event exists and get event details
	_, err = h.eventService.GetEventByID(eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	// Calculate total amount and validate ticket availability
	var totalAmount float64
	for _, ticketDetail := range req.TicketDetails {
		ticketTypeID, err := uuid.Parse(ticketDetail.TicketTypeID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ticket type ID"})
		}

		// Verify ticket type exists and has availability
		ticketType, err := h.ticketService.GetTicketTypeByID(ticketTypeID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Ticket type not found"})
		}

		if ticketType.SoldQuantity+ticketDetail.Quantity > ticketType.TotalQuantity {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Not enough tickets available for %s", ticketType.Name),
			})
		}

		totalAmount += ticketDetail.Price * float64(ticketDetail.Quantity)
	}

	// Generate payment reference
	reference := fmt.Sprintf("motiv_%s_%s_%d", req.EventID, userID.String()[:8], time.Now().Unix())

	// Create payment record
	payment := &models.Payment{
		EventID:   eventID,
		UserID:    userID,
		Amount:    totalAmount,
		Currency:  "NGN",
		Status:    models.PaymentPending,
		Method:    models.Card,
		Reference: reference,
	}

	err = h.paymentService.CreatePayment(payment)
	if err != nil {
		log.Printf("Error creating payment record: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create payment"})
	}

	// Return payment initiation response
	response := models.PaymentInitiationResponse{
		Reference:   reference,
		Amount:      int64(totalAmount * 100), // Convert to kobo
		PaystackURL: "https://checkout.paystack.com",
		PublicKey:   os.Getenv("PAYSTACK_PUBLIC_KEY"),
		Email:       req.Email,
		Currency:    "NGN",
	}

	return c.JSON(response)
}

// POST /api/v1/payments/webhook
func (h *PaymentHandler) PaymentWebhook(c *fiber.Ctx) error {
	log.Printf("Received webhook from Paystack")

	// Verify Paystack signature
	signature := c.Get("x-paystack-signature")
	if signature == "" {
		log.Printf("Missing Paystack signature")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing signature"})
	}

	body := c.Body()
	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")

	if secretKey == "" {
		log.Printf("PAYSTACK_SECRET_KEY not configured")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Server configuration error"})
	}

	// Verify signature
	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		log.Printf("Invalid Paystack signature. Expected: %s, Got: %s", expectedSignature, signature)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid signature"})
	}

	log.Printf("Webhook signature verified successfully")

	var webhookEvent models.PaystackWebhookEvent
	if err := json.Unmarshal(body, &webhookEvent); err != nil {
		log.Printf("Error parsing webhook payload: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	log.Printf("Webhook event type: %s, Reference: %s", webhookEvent.Event, webhookEvent.Data.Reference)

	// Handle different webhook events
	switch webhookEvent.Event {
	case "charge.success":
		err := h.handleSuccessfulPayment(webhookEvent)
		if err != nil {
			log.Printf("Error handling successful payment: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process payment"})
		}
	case "charge.failed":
		err := h.handleFailedPayment(webhookEvent)
		if err != nil {
			log.Printf("Error handling failed payment: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process payment failure"})
		}
	default:
		log.Printf("Unhandled webhook event: %s", webhookEvent.Event)
	}

	return c.JSON(fiber.Map{"message": "Webhook processed successfully"})
}

// GET /api/v1/payments/webhook/test - Test endpoint to verify webhook is reachable
func (h *PaymentHandler) TestWebhook(c *fiber.Ctx) error {
	log.Printf("Webhook test endpoint accessed from IP: %s", c.IP())
	return c.JSON(fiber.Map{
		"message":   "Webhook endpoint is reachable",
		"timestamp": time.Now().Format(time.RFC3339),
		"ip":        c.IP(),
	})
}

func (h *PaymentHandler) handleSuccessfulPayment(event models.PaystackWebhookEvent) error {
	log.Printf("Processing successful payment for reference: %s", event.Data.Reference)

	// Update payment status
	err := h.paymentService.UpdatePaymentStatus(event.Data.Reference, models.PaymentCompleted, "")
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Create tickets for the user
	eventID, err := uuid.Parse(event.Data.Metadata.EventID)
	if err != nil {
		return fmt.Errorf("invalid event ID in metadata: %w", err)
	}

	// Get event details for email
	eventDetails, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		return fmt.Errorf("failed to get event details: %w", err)
	}

	// Get host details for email
	host, err := h.userService.GetUserByID(eventDetails.HostID)
	if err != nil {
		return fmt.Errorf("failed to get host details: %w", err)
	}

	// Find user by email (assuming the customer email matches user email)
	userID, err := h.paymentService.GetUserIDByEmail(event.Data.Customer.Email)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}

	// Get user details for email
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user details: %w", err)
	}

	// Create tickets for each ticket type
	var ticketsCreated []*models.Ticket
	attendeeIndex := 0

	// Default to primary attendee if no additional attendees data available
	attendees := []struct {
		FullName string
		Email    string
		Phone    string
	}{{
		FullName: event.Data.Metadata.AttendeeData.FullName,
		Email:    event.Data.Metadata.AttendeeData.Email,
		Phone:    event.Data.Metadata.AttendeeData.Phone,
	}}

	log.Printf("Creating tickets for %d ticket types", len(event.Data.Metadata.TicketDetails))

	for _, ticketDetail := range event.Data.Metadata.TicketDetails {
		ticketTypeID, err := uuid.Parse(ticketDetail.TicketTypeID)
		if err != nil {
			log.Printf("Invalid ticket type ID: %s", ticketDetail.TicketTypeID)
			continue
		}

		log.Printf("Creating %d tickets for ticket type: %s", ticketDetail.Quantity, ticketDetail.TicketTypeName)

		for i := 0; i < ticketDetail.Quantity; i++ {
			// Cycle through attendees if we have more tickets than attendees
			currentAttendee := attendees[attendeeIndex%len(attendees)]

			ticket := &models.Ticket{
				EventID:          eventID,
				UserID:           userID,
				TicketTypeID:     ticketTypeID,
				PaymentReference: event.Data.Reference,
				AttendeeFullName: currentAttendee.FullName,
				AttendeeEmail:    currentAttendee.Email,
				AttendeePhone:    currentAttendee.Phone,
				Quantity:         1, // Each ticket is for one person
			}

			err = h.ticketService.CreateTicketWithQR(ticket)
			if err != nil {
				log.Printf("Failed to create ticket: %v", err)
				continue
			}

			log.Printf("Successfully created ticket %s for attendee %s", ticket.ID.String(), currentAttendee.FullName)
			ticketsCreated = append(ticketsCreated, ticket)
			attendeeIndex++
		}

		// Update ticket type sold quantity
		err = h.ticketService.UpdateSoldQuantity(ticketTypeID, ticketDetail.Quantity)
		if err != nil {
			log.Printf("Failed to update sold quantity: %v", err)
		}
	}

	log.Printf("Created %d tickets total", len(ticketsCreated))

	// Send email notifications for each ticket created
	for _, ticket := range ticketsCreated {
		// Send ticket confirmation email to customer
		if err := h.emailService.SendTicketConfirmation(ticket, eventDetails, user); err != nil {
			log.Printf("Failed to send ticket confirmation email: %v", err)
			// Don't fail the entire operation if email fails
		}

		// Send notification email to host
		if err := h.emailService.SendHostNotification(ticket, eventDetails, user, host); err != nil {
			log.Printf("Failed to send host notification email: %v", err)
			// Don't fail the entire operation if email fails
		}
	}

	log.Printf("Webhook processing completed successfully for reference: %s", event.Data.Reference)
	return nil
}

func (h *PaymentHandler) handleFailedPayment(event models.PaystackWebhookEvent) error {
	return h.paymentService.UpdatePaymentStatus(event.Data.Reference, models.PaymentFailed, event.Data.Message)
}

// POST /api/v1/payments/simulate-success - For testing without webhooks
func (h *PaymentHandler) SimulatePaymentSuccess(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	// Validate that the user exists
	userDetails, err := h.userService.GetUserByID(userID)
	if err != nil {
		log.Printf("User with ID %s not found: %v", userID, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	var req struct {
		Reference     string                       `json:"reference"`
		EventID       string                       `json:"eventId"`
		AttendeeData  models.AttendeeDataRequest   `json:"attendeeData"`
		Attendees     []models.AttendeeDataRequest `json:"attendees"`
		TicketDetails []models.TicketDetailRequest `json:"ticketDetails"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Log the request data for debugging
	log.Printf("Simulate success request - Reference: %s, EventID: %s", req.Reference, req.EventID)

	// Update payment status
	err = h.paymentService.UpdatePaymentStatus(req.Reference, models.PaymentCompleted, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update payment status"})
	}

	// Create tickets
	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		log.Printf("Invalid event ID format: %s", req.EventID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	log.Printf("Creating tickets for event ID: %s", eventID.String())

	// Get event details for email
	eventDetails, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		log.Printf("Event %s not found when creating ticket: %v", eventID.String(), err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Event not found"})
	}

	// Get host details for email
	host, err := h.userService.GetUserByID(eventDetails.HostID)
	if err != nil {
		log.Printf("Host with ID %s not found: %v", eventDetails.HostID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Host not found"})
	}

	// Use all attendees if provided, otherwise use primary attendee data
	attendees := req.Attendees
	if len(attendees) == 0 {
		attendees = []models.AttendeeDataRequest{req.AttendeeData}
	}

	var ticketsCreated []*models.Ticket
	attendeeIndex := 0
	for _, ticketDetail := range req.TicketDetails {
		ticketTypeID, err := uuid.Parse(ticketDetail.TicketTypeID)
		if err != nil {
			continue
		}

		// Create individual tickets for each quantity, assigning to different attendees
		for i := 0; i < ticketDetail.Quantity; i++ {
			// Cycle through attendees if we have more tickets than attendees
			currentAttendee := attendees[attendeeIndex%len(attendees)]

			ticket := &models.Ticket{
				EventID:          eventID,
				UserID:           userID,
				TicketTypeID:     ticketTypeID,
				PaymentReference: req.Reference,
				AttendeeFullName: currentAttendee.FullName,
				AttendeeEmail:    currentAttendee.Email,
				AttendeePhone:    currentAttendee.Phone,
				Quantity:         1, // Each ticket is for one person
			}

			log.Printf("Creating ticket for event %s, user %s, attendee %s", eventID.String(), userID.String(), currentAttendee.FullName)

			err = h.ticketService.CreateTicketWithQR(ticket)
			if err != nil {
				log.Printf("Failed to create ticket: %v", err)
				continue
			}

			log.Printf("Successfully created ticket %s for event %s", ticket.ID.String(), eventID.String())
			ticketsCreated = append(ticketsCreated, ticket)

			attendeeIndex++
		}

		// Update ticket type sold quantity
		err = h.ticketService.UpdateSoldQuantity(ticketTypeID, ticketDetail.Quantity)
		if err != nil {
			log.Printf("Failed to update sold quantity: %v", err)
		}
	}

	// Send email notifications for each ticket created
	for _, ticket := range ticketsCreated {
		// Send ticket confirmation email to customer
		if err := h.emailService.SendTicketConfirmation(ticket, eventDetails, userDetails); err != nil {
			log.Printf("Failed to send ticket confirmation email: %v", err)
			// Don't fail the entire operation if email fails
		}

		// Send notification email to host
		if err := h.emailService.SendHostNotification(ticket, eventDetails, userDetails, host); err != nil {
			log.Printf("Failed to send host notification email: %v", err)
			// Don't fail the entire operation if email fails
		}
	}

	return c.JSON(fiber.Map{"message": "Payment simulated and tickets created successfully"})
}
