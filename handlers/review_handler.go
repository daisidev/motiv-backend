package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
	"github.com/google/uuid"
)

type ReviewHandler struct {
	reviewService services.ReviewService
}

func NewReviewHandler(reviewService services.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

type CreateReviewRequest struct {
	EventID uuid.UUID `json:"eventId" validate:"required"`
	Rating  int       `json:"rating" validate:"required,min=1,max=5"`
	Comment string    `json:"comment"`
}

// POST /api/v1/reviews
func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}
	
	var req CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	review := &models.Review{
		EventID: req.EventID,
		UserID:  userID,
		Rating:  req.Rating,
		Comment: req.Comment,
	}
	
	err := h.reviewService.CreateReview(review)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	return c.Status(201).JSON(fiber.Map{
		"data": review,
	})
}

// GET /api/v1/events/:id/reviews
func (h *ReviewHandler) GetEventReviews(c *fiber.Ctx) error {
	eventIDStr := c.Params("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid event ID",
		})
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	reviews, err := h.reviewService.GetEventReviews(eventID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get reviews",
		})
	}
	
	// Get rating stats
	stats, err := h.reviewService.GetEventRatingStats(eventID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get rating stats",
		})
	}
	
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"reviews": reviews,
			"stats":   stats,
		},
	})
}

// GET /api/v1/hosts/me/reviews
func (h *ReviewHandler) GetHostReviews(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user ID"})
	}
	
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	
	reviews, err := h.reviewService.GetHostReviews(userID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get reviews",
		})
	}
	
	// Get rating stats
	stats, err := h.reviewService.GetHostRatingStats(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get rating stats",
		})
	}
	
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"reviews": reviews,
			"stats":   stats,
		},
	})
}

// PUT /api/v1/reviews/:id
func (h *ReviewHandler) UpdateReview(c *fiber.Ctx) error {
	reviewIDStr := c.Params("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid review ID",
		})
	}
	
	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	
	err = h.reviewService.UpdateReview(reviewID, updates)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "Review updated successfully",
	})
}

// DELETE /api/v1/reviews/:id
func (h *ReviewHandler) DeleteReview(c *fiber.Ctx) error {
	reviewIDStr := c.Params("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid review ID",
		})
	}
	
	err = h.reviewService.DeleteReview(reviewID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete review",
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "Review deleted successfully",
	})
}

// POST /api/v1/reviews/:id/helpful
func (h *ReviewHandler) MarkReviewHelpful(c *fiber.Ctx) error {
	reviewIDStr := c.Params("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid review ID",
		})
	}
	
	err = h.reviewService.MarkReviewHelpful(reviewID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to mark review as helpful",
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "Review marked as helpful",
	})
}