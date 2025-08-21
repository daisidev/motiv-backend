package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Check required environment variables
	requiredVars := []string{"FRONTEND_URL"}
	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Required environment variable %s is not set", envVar)
		}
	}

	fmt.Println("🧪 Testing Email Service...")
	fmt.Println("Using Zoho SMTP email service")

	// Create email service
	emailService := services.NewZohoEmailService()

	// Create test data
	event := &models.Event{
		ID:          uuid.New(),
		Title:       "Test Event - Email Service Test",
		Description: "This is a test event to verify email functionality",
		StartDate:   time.Now().AddDate(0, 0, 7), // 1 week from now
		StartTime:   "19:00",
		EndTime:     "22:00",
		Location:    "Test Venue, 123 Test Street, Test City",
		HostID:      uuid.New(),
	}

	user := &models.User{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "user@example.com", // Change this to your email for testing
	}

	host := &models.User{
		ID:    uuid.New(),
		Name:  "Jane Smith",
		Email: "host@example.com", // Change this to your email for testing
	}
	ticket := &models.Ticket{
		ID:               uuid.New(),
		EventID:          event.ID,
		UserID:           user.ID,
		AttendeeFullName: "John Doe",
		AttendeeEmail:    "attendee@example.com", // Change this to your email for testing
		AttendeePhone:    "+1234567890",
		PaymentReference: "TEST_REF_" + uuid.New().String()[:8],
		Quantity:         1,
		QRCode:           "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==", // Minimal test QR
	}

	fmt.Println("\n📧 Sending test ticket confirmation email...")
	if err := emailService.SendTicketConfirmation(ticket, event, user); err != nil {
		log.Printf("❌ Failed to send ticket confirmation email: %v", err)
	} else {
		fmt.Println("✅ Ticket confirmation email sent successfully!")
	}

	fmt.Println("\n📨 Sending test host notification email...")
	if err := emailService.SendHostNotification(ticket, event, user, host); err != nil {
		log.Printf("❌ Failed to send host notification email: %v", err)
	} else {
		fmt.Println("✅ Host notification email sent successfully!")
	}

	fmt.Println("\n🎉 Email service test completed!")
	fmt.Println("Check your email inbox (and spam folder) for the test emails.")
	fmt.Println("Note: Update the email addresses in this script to receive actual test emails.")
}
