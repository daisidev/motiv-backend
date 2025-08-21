package services

import (
	"fmt"
	"log"
	"os"

	"github.com/hidenkeys/motiv-backend/models"
)

// MockEmailService is a mock implementation that logs emails instead of sending them
type MockEmailService struct{}

func NewMockEmailService() EmailService {
	return &MockEmailService{}
}

func (m *MockEmailService) SendTicketConfirmation(ticket *models.Ticket, event *models.Event, user *models.User) error {
	log.Printf("MOCK EMAIL: Ticket confirmation sent to %s for event %s", ticket.AttendeeEmail, event.Title)
	return nil
}

func (m *MockEmailService) SendHostNotification(ticket *models.Ticket, event *models.Event, user *models.User, host *models.User) error {
	log.Printf("MOCK EMAIL: Host notification sent to %s for event %s", host.Email, event.Title)
	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(user *models.User, resetToken string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)

	log.Printf("MOCK EMAIL: Password reset email for user %s (%s)", user.Name, user.Email)
	log.Printf("Reset link: %s", resetLink)
	log.Printf("Token expires in 1 hour")

	// In a real scenario, you would send this email
	fmt.Printf("\n=== PASSWORD RESET EMAIL ===\n")
	fmt.Printf("To: %s\n", user.Email)
	fmt.Printf("Subject: Reset Your Password - Motiv Events\n")
	fmt.Printf("Reset Link: %s\n", resetLink)
	fmt.Printf("===========================\n\n")

	return nil
}

func (m *MockEmailService) SendWelcomeEmail(user *models.User) error {
	log.Printf("MOCK EMAIL: Welcome email sent to %s (%s)", user.Name, user.Email)

	fmt.Printf("\n=== WELCOME EMAIL ===\n")
	fmt.Printf("To: %s\n", user.Email)
	fmt.Printf("Subject: Welcome to Motiv Events!\n")
	fmt.Printf("Welcome message for user: %s (Username: %s)\n", user.Name, user.Username)
	fmt.Printf("====================\n\n")

	return nil
}
