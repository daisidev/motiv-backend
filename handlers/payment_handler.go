package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler(paymentService services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
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

// POST /api/v1/payments/webhook
func (h *PaymentHandler) PaymentWebhook(c *fiber.Ctx) error {
	var req UpdatePaymentStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	// Convert string status to enum
	var status models.PaymentStatus
	switch req.Status {
	case "completed":
		status = models.PaymentCompleted
	case "failed":
		status = models.PaymentFailed
	case "pending":
		status = models.PaymentPending
	case "refunded":
		status = models.PaymentRefunded
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid payment status",
		})
	}
	
	err := h.paymentService.UpdatePaymentStatus(req.Reference, status, req.FailureReason)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update payment status",
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "Payment status updated successfully",
	})
}