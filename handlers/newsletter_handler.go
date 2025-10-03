package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hidenkeys/motiv-backend/models"
	"log"
)

type NewsletterHandler struct {
	userService models.UserService
}

func NewNewsletterHandler(userService models.UserService) *NewsletterHandler {
	return &NewsletterHandler{
		userService: userService,
	}
}

// Subscribe handles newsletter subscription requests
func (h *NewsletterHandler) Subscribe(c *fiber.Ctx) error {
	var request struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if request.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	log.Printf("📧 Newsletter subscription request for email: %s", request.Email)

	// Check if user exists with this email
	user, err := h.userService.GetUserByEmail(request.Email)
	if err == nil && user != nil {
		// User exists, update their newsletter subscription
		user.NewsletterSubscribed = true
		if updateErr := h.userService.UpdateUser(user); updateErr != nil {
			log.Printf("❌ Failed to update existing user newsletter subscription: %v", updateErr)
		} else {
			log.Printf("✅ Updated existing user newsletter subscription for: %s", request.Email)
		}
	} else {
		// User doesn't exist, just log the subscription request
		// In a real implementation, you might want to store this in a separate newsletter_subscribers table
		log.Printf("📋 Newsletter subscription logged for non-user email: %s", request.Email)
	}

	return c.JSON(fiber.Map{
		"message": "Successfully subscribed to newsletter",
		"email":   request.Email,
	})
}

// Unsubscribe handles newsletter unsubscription requests
func (h *NewsletterHandler) Unsubscribe(c *fiber.Ctx) error {
	var request struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if request.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	log.Printf("📧 Newsletter unsubscription request for email: %s", request.Email)

	// Check if user exists with this email
	user, err := h.userService.GetUserByEmail(request.Email)
	if err == nil && user != nil {
		// User exists, update their newsletter subscription
		user.NewsletterSubscribed = false
		if updateErr := h.userService.UpdateUser(user); updateErr != nil {
			log.Printf("❌ Failed to update user newsletter unsubscription: %v", updateErr)
		} else {
			log.Printf("✅ Updated user newsletter unsubscription for: %s", request.Email)
		}
	} else {
		log.Printf("📋 Newsletter unsubscription logged for non-user email: %s", request.Email)
	}

	return c.JSON(fiber.Map{
		"message": "Successfully unsubscribed from newsletter",
		"email":   request.Email,
	})
}