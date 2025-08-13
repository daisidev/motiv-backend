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
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService services.PaymentService
	ticketService  services.TicketService
	eventService   services.EventService
}

func NewPaymentHandler(paymentService services.PaymentService, ticketService services.TicketService, eventService services.EventService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		ticketService:  ticketService,
		eventService:   eventService,
	}
}

// GET /api/v1/hosts/me/payments/earnings
func (h *PaymentHandler) GetHostEarnings(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	
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
	userID := c.Locals("userID").(uuid.UUID)
	
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
	userID := c.Locals("userID").(uuid.UUID)
	
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
	event, err := h.eventService.GetEventByID(eventID)
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
	// Verify Paystack signature
	signature := c.Get("x-paystack-signature")
	if signature == "" {
		log.Printf("Missing Paystack signature")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing signature"})
	}

	body := c.Body()
	secretKey := os.Getenv("PAYSTACK_SECRET_KEY")
	
	// Verify signature
	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		log.Printf("Invalid Paystack signature")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid signature"})
	}

	var webhookEvent models.PaystackWebhookEvent
	if err := json.Unmarshal(body, &webhookEvent); err != nil {
		log.Printf("Error parsing webhook payload: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

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

func (h *PaymentHandler) handleSuccessfulPayment(event models.PaystackWebhookEvent) error {
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

	// Find user by email (assuming the customer email matches user email)
	userID, err := h.paymentService.GetUserIDByEmail(event.Data.Customer.Email)
	if err != nil {
		return fmt.Errorf("failed to find user by email: %w", err)
	}

	// Create tickets for each ticket type
	for _, ticketDetail := range event.Data.Metadata.TicketDetails {
		ticketTypeID, err := uuid.Parse(ticketDetail.TicketTypeID)
		if err != nil {
			return fmt.Errorf("invalid ticket type ID: %w", err)
		}

		for i := 0; i < ticketDetail.Quantity; i++ {
			ticket := &models.Ticket{
				EventID:          eventID,
				UserID:           userID,
				TicketTypeID:     ticketTypeID,
				PaymentReference: event.Data.Reference,
				AttendeeFullName: event.Data.Metadata.AttendeeData.FullName,
				AttendeeEmail:    event.Data.Metadata.AttendeeData.Email,
				AttendeePhone:    event.Data.Metadata.AttendeeData.Phone,
				Quantity:         ticketDetail.Quantity,
			}

			err = h.ticketService.CreateTicketWithQR(ticket)
			if err != nil {
				return fmt.Errorf("failed to create ticket: %w", err)
			}
		}

		// Update ticket type sold quantity
		err = h.ticketService.UpdateSoldQuantity(ticketTypeID, ticketDetail.Quantity)
		if err != nil {
			return fmt.Errorf("failed to update sold quantity: %w", err)
		}
	}

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

	var req struct {
		Reference     string `json:"reference"`
		EventID       string `json:"eventId"`
		AttendeeData  models.AttendeeDataRequest `json:"attendeeData"`
		TicketDetails []models.TicketDetailRequest `json:"ticketDetails"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update payment status
	err = h.paymentService.UpdatePaymentStatus(req.Reference, models.PaymentCompleted, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update payment status"})
	}

	// Create tickets
	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	for _, ticketDetail := range req.TicketDetails {
		ticketTypeID, err := uuid.Parse(ticketDetail.TicketTypeID)
		if err != nil {
			continue
		}

		for i := 0; i < ticketDetail.Quantity; i++ {
			ticket := &models.Ticket{
				EventID:          eventID,
				UserID:           userID,
				TicketTypeID:     ticketTypeID,
				PaymentReference: req.Reference,
				AttendeeFullName: req.AttendeeData.FullName,
				AttendeeEmail:    req.AttendeeData.Email,
				AttendeePhone:    req.AttendeeData.Phone,
				Quantity:         ticketDetail.Quantity,
			}

			err = h.ticketService.CreateTicketWithQR(ticket)
			if err != nil {
				log.Printf("Failed to create ticket: %v", err)
				continue
			}
		}

		// Update ticket type sold quantity
		err = h.ticketService.UpdateSoldQuantity(ticketTypeID, ticketDetail.Quantity)
		if err != nil {
			log.Printf("Failed to update sold quantity: %v", err)
		}
	}

	return c.JSON(fiber.Map{"message": "Payment simulated and tickets created successfully"})
}