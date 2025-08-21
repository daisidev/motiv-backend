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
		log.Printf("❌ PAYMENT INIT ERROR: Failed to parse user ID: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	log.Printf("💳 PAYMENT INITIATION: Starting payment initiation for user: %s", userID.String())

	// Validate that the user exists
	userDetails, err := h.userService.GetUserByID(userID)
	if err != nil {
		log.Printf("❌ PAYMENT INIT ERROR: User with ID %s not found: %v", userID, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}
	log.Printf("👤 USER VERIFIED: User %s (%s) found for payment initiation", userDetails.Email, userDetails.Name)

	var req models.PaymentInitiationRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("❌ PAYMENT INIT ERROR: Error parsing payment initiation request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	log.Printf("📋 PAYMENT REQUEST: EventID=%s, Email=%s, TicketTypes=%d", req.EventID, req.Email, len(req.TicketDetails))
	log.Printf("👥 ATTENDEE INFO: %s (%s, %s)", req.AttendeeData.FullName, req.AttendeeData.Email, req.AttendeeData.Phone)

	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		log.Printf("❌ PAYMENT INIT ERROR: Invalid event ID format: %s", req.EventID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	// Verify event exists and get event details
	eventDetails, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		log.Printf("❌ PAYMENT INIT ERROR: Event %s not found: %v", eventID.String(), err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}
	log.Printf("📅 EVENT VERIFIED: Event '%s' found for payment", eventDetails.Title)

	// Calculate total amount and validate ticket availability
	var totalAmount float64
	log.Printf("💰 CALCULATING TOTAL: Starting ticket validation and amount calculation")

	for i, ticketDetail := range req.TicketDetails {
		log.Printf("🎫 TICKET VALIDATION %d/%d: Type=%s, Quantity=%d, Price=%.2f",
			i+1, len(req.TicketDetails), ticketDetail.TicketTypeID, ticketDetail.Quantity, ticketDetail.Price)

		ticketTypeID, err := uuid.Parse(ticketDetail.TicketTypeID)
		if err != nil {
			log.Printf("❌ PAYMENT INIT ERROR: Invalid ticket type ID: %s", ticketDetail.TicketTypeID)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ticket type ID"})
		}

		// Verify ticket type exists and has availability
		ticketType, err := h.ticketService.GetTicketTypeByID(ticketTypeID)
		if err != nil {
			log.Printf("❌ PAYMENT INIT ERROR: Ticket type %s not found: %v", ticketTypeID.String(), err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Ticket type not found"})
		}

		log.Printf("🎫 TICKET TYPE VERIFIED: %s - Available: %d, Sold: %d, Requesting: %d",
			ticketType.Name, ticketType.TotalQuantity-ticketType.SoldQuantity, ticketType.SoldQuantity, ticketDetail.Quantity)

		if ticketType.SoldQuantity+ticketDetail.Quantity > ticketType.TotalQuantity {
			log.Printf("❌ PAYMENT INIT ERROR: Not enough tickets available for %s. Available: %d, Requested: %d",
				ticketType.Name, ticketType.TotalQuantity-ticketType.SoldQuantity, ticketDetail.Quantity)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Not enough tickets available for %s", ticketType.Name),
			})
		}

		subtotal := ticketDetail.Price * float64(ticketDetail.Quantity)
		totalAmount += subtotal
		log.Printf("💰 SUBTOTAL: %s x %d = %.2f NGN", ticketType.Name, ticketDetail.Quantity, subtotal)
	}

	log.Printf("💰 TOTAL AMOUNT: %.2f NGN (%.0f kobo)", totalAmount, totalAmount*100)

	// Generate payment reference
	reference := fmt.Sprintf("motiv_%s_%s_%d", req.EventID, userID.String()[:8], time.Now().Unix())
	log.Printf("🔗 PAYMENT REFERENCE: Generated reference: %s", reference)

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

	log.Printf("💾 PAYMENT RECORD: Creating payment record with status PENDING")
	err = h.paymentService.CreatePayment(payment)
	if err != nil {
		log.Printf("❌ PAYMENT INIT ERROR: Error creating payment record: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create payment"})
	}
	log.Printf("✅ PAYMENT RECORD: Payment record created successfully with ID: %s", payment.ID.String())

	// Return payment initiation response
	response := models.PaymentInitiationResponse{
		Reference:   reference,
		Amount:      int64(totalAmount * 100), // Convert to kobo
		PaystackURL: "https://checkout.paystack.com",
		PublicKey:   os.Getenv("PAYSTACK_PUBLIC_KEY"),
		Email:       req.Email,
		Currency:    "NGN",
	}

	log.Printf("🚀 PAYMENT INITIATED: Returning payment initiation response for %.2f NGN", totalAmount)
	log.Printf("⏳ AWAITING WEBHOOK: Payment %s is now pending webhook confirmation", reference)
	return c.JSON(response)
}

// POST /api/v1/payments/webhook
func (h *PaymentHandler) PaymentWebhook(c *fiber.Ctx) error {
	log.Printf("🔔 WEBHOOK RECEIVED: Payment webhook called from IP: %s", c.IP())
	log.Printf("🔔 WEBHOOK HEADERS: %+v", c.GetReqHeaders())

	// Verify Paystack signature
	signature := c.Get("x-paystack-signature")
	if signature == "" {
		log.Printf("❌ WEBHOOK ERROR: Missing Paystack signature")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing signature"})
	}

	body := c.Body()
	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")

	if secretKey == "" {
		log.Printf("❌ WEBHOOK ERROR: PAYSTACK_SECRET_KEY not configured")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Server configuration error"})
	}

	log.Printf("🔐 WEBHOOK VERIFICATION: Verifying signature for payload length: %d bytes", len(body))

	// Verify signature
	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		log.Printf("❌ WEBHOOK ERROR: Invalid Paystack signature. Expected: %s, Got: %s", expectedSignature, signature)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid signature"})
	}

	log.Printf("✅ WEBHOOK VERIFICATION: Signature verified successfully")

	var webhookEvent models.PaystackWebhookEvent
	if err := json.Unmarshal(body, &webhookEvent); err != nil {
		log.Printf("❌ WEBHOOK ERROR: Error parsing webhook payload: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	log.Printf("🎯 WEBHOOK EVENT: Type=%s, Reference=%s, Amount=%v, Status=%s",
		webhookEvent.Event,
		webhookEvent.Data.Reference,
		webhookEvent.Data.Amount,
		webhookEvent.Data.Status)

	log.Printf("📧 WEBHOOK CUSTOMER: Email=%s", webhookEvent.Data.Customer.Email)
	log.Printf("📊 WEBHOOK METADATA: EventID=%s", webhookEvent.Data.Metadata.EventID)
	log.Printf("📊 WEBHOOK ATTENDEE: Name=%s, Email=%s, Phone=%s",
		webhookEvent.Data.Metadata.AttendeeData.FullName,
		webhookEvent.Data.Metadata.AttendeeData.Email,
		webhookEvent.Data.Metadata.AttendeeData.Phone)
	log.Printf("📊 WEBHOOK TICKETS: %d ticket types", len(webhookEvent.Data.Metadata.TicketDetails))
	for i, ticket := range webhookEvent.Data.Metadata.TicketDetails {
		log.Printf("📊 WEBHOOK TICKET %d: TypeID=%s, Name=%s, Quantity=%d, Price=%.2f",
			i+1, ticket.TicketTypeID, ticket.TicketTypeName, ticket.Quantity, ticket.Price)
	} // Handle different webhook events
	switch webhookEvent.Event {
	case "charge.success":
		log.Printf("💳 WEBHOOK PROCESSING: Handling successful payment for reference: %s", webhookEvent.Data.Reference)
		err := h.handleSuccessfulPayment(webhookEvent)
		if err != nil {
			log.Printf("❌ WEBHOOK ERROR: Error handling successful payment: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process payment"})
		}
		log.Printf("✅ WEBHOOK SUCCESS: Successfully processed payment for reference: %s", webhookEvent.Data.Reference)
	case "charge.failed":
		log.Printf("💔 WEBHOOK PROCESSING: Handling failed payment for reference: %s", webhookEvent.Data.Reference)
		err := h.handleFailedPayment(webhookEvent)
		if err != nil {
			log.Printf("❌ WEBHOOK ERROR: Error handling failed payment: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process payment failure"})
		}
		log.Printf("✅ WEBHOOK SUCCESS: Successfully processed payment failure for reference: %s", webhookEvent.Data.Reference)
	default:
		log.Printf("⚠️ WEBHOOK WARNING: Unhandled webhook event: %s for reference: %s", webhookEvent.Event, webhookEvent.Data.Reference)
	}

	log.Printf("🎉 WEBHOOK COMPLETE: Webhook processed successfully for event: %s, reference: %s", webhookEvent.Event, webhookEvent.Data.Reference)
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
	log.Printf("🚀 PAYMENT SUCCESS: Processing successful payment for reference: %s", event.Data.Reference)
	log.Printf("💰 PAYMENT DETAILS: Amount=%v, Currency=%s, Channel=%s", event.Data.Amount, event.Data.Currency, event.Data.Channel)

	// Update payment status
	err := h.paymentService.UpdatePaymentStatus(event.Data.Reference, models.PaymentCompleted, "")
	if err != nil {
		log.Printf("❌ PAYMENT ERROR: Failed to update payment status for reference %s: %v", event.Data.Reference, err)
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	log.Printf("✅ PAYMENT UPDATE: Payment status updated to completed for reference: %s", event.Data.Reference)

	// Validate event ID in metadata
	if event.Data.Metadata.EventID == "" {
		log.Printf("❌ PAYMENT ERROR: Missing event ID in metadata for reference: %s", event.Data.Reference)
		return fmt.Errorf("missing event ID in payment metadata")
	}

	// Create tickets for the user
	eventID, err := uuid.Parse(event.Data.Metadata.EventID)
	if err != nil {
		log.Printf("❌ PAYMENT ERROR: Invalid event ID format in metadata for reference %s: %v", event.Data.Reference, err)
		return fmt.Errorf("invalid event ID in metadata: %w", err)
	}
	log.Printf("🎫 TICKET CREATION: Creating tickets for event: %s", eventID.String())

	// Get event details for email
	eventDetails, err := h.eventService.GetEventByID(eventID)
	if err != nil {
		log.Printf("❌ PAYMENT ERROR: Failed to get event details for event %s: %v", eventID.String(), err)
		return fmt.Errorf("failed to get event details: %w", err)
	}
	log.Printf("📅 EVENT DETAILS: Found event '%s' for payment reference: %s", eventDetails.Title, event.Data.Reference)

	// Get host details for email
	host, err := h.userService.GetUserByID(eventDetails.HostID)
	if err != nil {
		log.Printf("❌ PAYMENT ERROR: Failed to get host details for host %s: %v", eventDetails.HostID.String(), err)
		return fmt.Errorf("failed to get host details: %w", err)
	}
	log.Printf("👤 HOST DETAILS: Found host '%s' for event: %s", host.Email, eventDetails.Title)

	// Find user by email (assuming the customer email matches user email)
	log.Printf("🔍 USER LOOKUP: Searching for user with email: %s", event.Data.Customer.Email)
	userID, err := h.paymentService.GetUserIDByEmail(event.Data.Customer.Email)
	if err != nil {
		log.Printf("❌ PAYMENT ERROR: Failed to find user by email %s: %v", event.Data.Customer.Email, err)
		log.Printf("💡 EMAIL MISMATCH: The webhook customer email (%s) doesn't match any user in database", event.Data.Customer.Email)
		log.Printf("💡 EMAIL MISMATCH: This could mean the payment email differs from the registered user email")
		return fmt.Errorf("failed to find user by email: %w", err)
	}
	log.Printf("👤 USER DETAILS: Found user ID %s for email: %s", userID.String(), event.Data.Customer.Email)

	// Get user details for email
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		log.Printf("❌ PAYMENT ERROR: Failed to get user details for user %s: %v", userID.String(), err)
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

	log.Printf("🎟️ TICKET PROCESSING: Creating tickets for %d ticket types", len(event.Data.Metadata.TicketDetails))
	log.Printf("👥 ATTENDEE INFO: Primary attendee: %s (%s)", attendees[0].FullName, attendees[0].Email)

	if len(event.Data.Metadata.TicketDetails) == 0 {
		log.Printf("⚠️ TICKET WARNING: No ticket details found in payment metadata for reference: %s", event.Data.Reference)
		return fmt.Errorf("no ticket details found in payment metadata")
	}

	for _, ticketDetail := range event.Data.Metadata.TicketDetails {
		ticketTypeID, err := uuid.Parse(ticketDetail.TicketTypeID)
		if err != nil {
			log.Printf("❌ TICKET ERROR: Invalid ticket type ID: %s for reference: %s", ticketDetail.TicketTypeID, event.Data.Reference)
			continue
		}

		log.Printf("🎫 TICKET CREATION: Creating %d tickets for ticket type: %s (ID: %s, Price: %v)",
			ticketDetail.Quantity,
			ticketDetail.TicketTypeName,
			ticketDetail.TicketTypeID,
			ticketDetail.Price)

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

			log.Printf("🎫 TICKET CREATING: Ticket %d/%d for attendee: %s", i+1, ticketDetail.Quantity, currentAttendee.FullName)

			err = h.ticketService.CreateTicketWithQR(ticket)
			if err != nil {
				log.Printf("❌ TICKET ERROR: Failed to create ticket %d for type %s: %v", i+1, ticketDetail.TicketTypeName, err)
				continue
			}

			log.Printf("✅ TICKET CREATED: Successfully created ticket %s for attendee %s", ticket.ID.String(), currentAttendee.FullName)
			ticketsCreated = append(ticketsCreated, ticket)
			attendeeIndex++
		}

		// Update ticket type sold quantity
		log.Printf("📊 UPDATING SALES: Updating sold quantity for ticket type %s by %d", ticketTypeID.String(), ticketDetail.Quantity)
		err = h.ticketService.UpdateSoldQuantity(ticketTypeID, ticketDetail.Quantity)
		if err != nil {
			log.Printf("❌ SALES ERROR: Failed to update sold quantity for ticket type %s: %v", ticketTypeID.String(), err)
		} else {
			log.Printf("✅ SALES UPDATED: Successfully updated sold quantity for ticket type %s", ticketTypeID.String())
		}
	}

	log.Printf("🎉 TICKETS CREATED: Created %d tickets total for payment reference: %s", len(ticketsCreated), event.Data.Reference)

	// Send email notifications for each ticket created
	log.Printf("📧 EMAIL NOTIFICATIONS: Starting email notifications for %d tickets", len(ticketsCreated))
	for i, ticket := range ticketsCreated {
		log.Printf("📧 EMAIL SENDING: Sending notifications for ticket %d/%d (ID: %s)", i+1, len(ticketsCreated), ticket.ID.String())
		log.Printf("📧 EMAIL DETAILS: Customer=%s (UserID: %s), Attendee=%s (%s)", user.Email, user.ID.String(), ticket.AttendeeFullName, ticket.AttendeeEmail)
		log.Printf("📧 EMAIL DETAILS: Host=%s (UserID: %s), Event=%s", host.Email, host.ID.String(), eventDetails.Title)

		// Send ticket confirmation email to customer
		log.Printf("📧 SENDING CUSTOMER EMAIL: To %s for ticket %s", ticket.AttendeeEmail, ticket.ID.String())
		if err := h.emailService.SendTicketConfirmation(ticket, eventDetails, user); err != nil {
			log.Printf("❌ EMAIL ERROR: Failed to send ticket confirmation email for ticket %s to %s: %v", ticket.ID.String(), ticket.AttendeeEmail, err)
			// Don't fail the entire operation if email fails
		} else {
			log.Printf("✅ EMAIL SENT: Ticket confirmation email sent successfully for ticket %s to %s", ticket.ID.String(), ticket.AttendeeEmail)
		}

		// Send notification email to host
		log.Printf("📧 SENDING HOST EMAIL: To %s for ticket %s", host.Email, ticket.ID.String())
		if err := h.emailService.SendHostNotification(ticket, eventDetails, user, host); err != nil {
			log.Printf("❌ EMAIL ERROR: Failed to send host notification email for ticket %s to %s: %v", ticket.ID.String(), host.Email, err)
			// Don't fail the entire operation if email fails
		} else {
			log.Printf("✅ EMAIL SENT: Host notification email sent successfully for ticket %s to %s", ticket.ID.String(), host.Email)
		}
	}

	log.Printf("🎉 WEBHOOK SUCCESS: Webhook processing completed successfully for reference: %s - Created %d tickets", event.Data.Reference, len(ticketsCreated))
	return nil
}

func (h *PaymentHandler) handleFailedPayment(event models.PaystackWebhookEvent) error {
	return h.paymentService.UpdatePaymentStatus(event.Data.Reference, models.PaymentFailed, event.Data.Message)
}

// POST /api/v1/payments/simulate-success - For testing without webhooks
func (h *PaymentHandler) SimulatePaymentSuccess(c *fiber.Ctx) error {
	log.Printf("🧪 SIMULATE PAYMENT: Payment simulation endpoint called - this should only be used for testing!")

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		log.Printf("❌ SIMULATE ERROR: Failed to parse user ID: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}

	log.Printf("🧪 SIMULATE INITIATED: User %s is simulating payment success", userID.String())

	// Validate that the user exists
	userDetails, err := h.userService.GetUserByID(userID)
	if err != nil {
		log.Printf("❌ SIMULATE ERROR: User with ID %s not found: %v", userID, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	log.Printf("👤 USER VERIFIED: User %s (%s) found for simulation", userDetails.Email, userDetails.Name)

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
