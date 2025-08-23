package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/hidenkeys/motiv-backend/config"
	"github.com/hidenkeys/motiv-backend/handlers"
	"github.com/hidenkeys/motiv-backend/middleware"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
	"github.com/hidenkeys/motiv-backend/services"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	// Connect to the database
	config.ConnectDatabase()
	config.MigrateDatabase()

	// Create repositories
	userRepo := repository.NewUserRepoPG(config.DB)
	eventRepo := repository.NewEventRepoPG(config.DB)
	ticketRepo := repository.NewTicketRepoPG(config.DB)
	wishlistRepo := repository.NewWishlistRepoPG(config.DB)
	reviewRepo := repository.NewReviewRepoPG(config.DB)
	paymentRepo := repository.NewPaymentRepoPG(config.DB)
	analyticsRepo := repository.NewAnalyticsRepoPG(config.DB)
	attendeeRepo := repository.NewAttendeeRepoPG(config.DB)

	// Create services
	userService := services.NewUserService(userRepo)
	eventService := services.NewEventService(eventRepo)
	ticketService := services.NewTicketService(ticketRepo, attendeeRepo)
	wishlistService := services.NewWishlistService(wishlistRepo)
	reviewService := services.NewReviewService(reviewRepo)
	paymentService := services.NewPaymentService(paymentRepo, userRepo)
	analyticsService := services.NewAnalyticsService(analyticsRepo, paymentRepo, attendeeRepo, reviewRepo)
	attendeeService := services.NewAttendeeService(attendeeRepo, ticketRepo)

	// Use Zoho email service
	var emailService services.EmailService
	log.Println("Using Zoho email service")
	emailService = services.NewZohoEmailService()

	// Create handlers
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	authHandler := handlers.NewAuthHandler(userService, emailService, jwtSecret)
	userHandler := handlers.NewUserHandler(userService, wishlistService, ticketService)
	eventHandler := handlers.NewEventHandler(eventService, ticketService)
	ticketHandler := handlers.NewTicketHandler(ticketService, eventService, userService, emailService)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	paymentHandler := handlers.NewPaymentHandler(paymentService, ticketService, eventService, userService, emailService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	attendeeHandler := handlers.NewAttendeeHandler(attendeeService, eventService)

	// Create Fiber app
	app := fiber.New()

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, http://localhost:3001, https://motiv-six.vercel.app, http://127.0.0.1:3000, https://motiv-six.vercel.app, https://motiv-alpha-seven.vercel.app",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
	app.Use(logger.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "MOTIV Backend is running",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/signup", authHandler.Signup)
	auth.Post("/login", authHandler.Login)
	auth.Post("/google", authHandler.GoogleAuth)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)

	// User routes
	user := api.Group("/users")
	user.Use(middleware.AuthRequired(jwtSecret))
	user.Get("/me", userHandler.GetMe)
	user.Put("/me", userHandler.UpdateMe)
	user.Get("/me/tickets", userHandler.GetMyTickets)
	user.Get("/me/tickets/:id", userHandler.GetMyTicket)
	user.Get("/me/tickets/debug", userHandler.GetMyTicketsDebug)
	user.Get("/me/wishlist", userHandler.GetMyWishlist)
	user.Get("/me/wishlist/check", userHandler.CheckWishlistStatus)
	user.Post("/me/wishlist", userHandler.AddToMyWishlist)
	user.Delete("/me/wishlist", userHandler.RemoveFromMyWishlist)

	// Event routes
	event := api.Group("/events")
	event.Get("/", eventHandler.GetAllEvents)
	event.Get("/suggestions", eventHandler.GetSearchSuggestions)
	event.Get("/:id", eventHandler.GetEventByID)
	event.Get("/:id/reviews", reviewHandler.GetEventReviews)
	event.Get("/:id/analytics", analyticsHandler.GetEventAnalytics)
	event.Get("/:id/revenue", paymentHandler.GetEventRevenue)
	event.Post("/:id/view", analyticsHandler.RecordEventView) // Optional auth

	// Host routes
	host := api.Group("/hosts")
	host.Use(middleware.AuthRequired(jwtSecret))
	host.Use(middleware.RoleRequired(models.HostRole, models.AdminRole, models.SuperhostRole))

	// Host events
	host.Get("/me/events", eventHandler.GetMyEvents)
	host.Post("/me/events", eventHandler.CreateEvent)
	host.Put("/me/events/:id", eventHandler.UpdateEvent)
	host.Delete("/me/events/:id", eventHandler.DeleteEvent)

	// Host analytics
	host.Get("/me/analytics/dashboard", analyticsHandler.GetHostDashboard)
	host.Get("/me/analytics/revenue", analyticsHandler.GetMonthlyRevenue)

	// Host reviews
	host.Get("/me/reviews", reviewHandler.GetHostReviews)

	// Host payments
	host.Get("/me/payments/earnings", paymentHandler.GetHostEarnings)
	host.Get("/me/payments/payouts", paymentHandler.GetHostPayouts)
	host.Get("/me/payments/pending", paymentHandler.GetPendingPayouts)

	// Host attendees
	host.Get("/me/attendees", attendeeHandler.GetHostAttendees)
	host.Get("/me/attendees/export", attendeeHandler.ExportHostAttendees)
	host.Get("/me/events/:eventId/attendees", attendeeHandler.GetEventAttendees)
	host.Post("/me/attendees/checkin", attendeeHandler.CheckInAttendee)

	// Review routes
	review := api.Group("/reviews")
	review.Use(middleware.AuthRequired(jwtSecret))
	review.Post("/", reviewHandler.CreateReview)
	review.Put("/:id", reviewHandler.UpdateReview)
	review.Delete("/:id", reviewHandler.DeleteReview)
	review.Post("/:id/helpful", reviewHandler.MarkReviewHelpful)

	// Payment routes
	payment := api.Group("/payments")
	payment.Post("/initiate", middleware.AuthRequired(jwtSecret), paymentHandler.InitiatePayment)
	payment.Post("/webhook", paymentHandler.PaymentWebhook)  // No auth required for webhook
	payment.Get("/webhook/test", paymentHandler.TestWebhook) // Test endpoint to verify webhook is reachable
	// payment.Post("/simulate-success", middleware.AuthRequired(jwtSecret), paymentHandler.SimulatePaymentSuccess) // For testing without webhooks - DISABLED for production

	// Ticket routes
	ticket := api.Group("/tickets")
	ticket.Use(middleware.AuthRequired(jwtSecret))
	ticket.Post("/purchase", ticketHandler.PurchaseTicket)
	ticket.Post("/rsvp", ticketHandler.RSVPFreeEvent)

	// Start server
	log.Fatal(app.Listen(":8080"))
}
