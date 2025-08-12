
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
	ticketService := services.NewTicketService(ticketRepo)
	wishlistService := services.NewWishlistService(wishlistRepo)
	reviewService := services.NewReviewService(reviewRepo)
	paymentService := services.NewPaymentService(paymentRepo)
	analyticsService := services.NewAnalyticsService(analyticsRepo, paymentRepo, attendeeRepo, reviewRepo)

	// Create handlers
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	authHandler := handlers.NewAuthHandler(userService, jwtSecret)
	userHandler := handlers.NewUserHandler(userService, wishlistService, ticketService)
	eventHandler := handlers.NewEventHandler(eventService)
	ticketHandler := handlers.NewTicketHandler(ticketService)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	// Create Fiber app
	app := fiber.New()

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, http://127.0.0.1:3000",
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

	// User routes
	user := api.Group("/users")
	user.Use(middleware.AuthRequired(jwtSecret))
	user.Get("/me", userHandler.GetMe)
	user.Put("/me", userHandler.UpdateMe)
	user.Get("/me/tickets", userHandler.GetMyTickets)
	user.Get("/me/wishlist", userHandler.GetMyWishlist)
	user.Post("/me/wishlist", userHandler.AddToMyWishlist)
	user.Delete("/me/wishlist", userHandler.RemoveFromMyWishlist)

	// Event routes
	event := api.Group("/events")
	event.Get("/", eventHandler.GetAllEvents)
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

	// Review routes
	review := api.Group("/reviews")
	review.Use(middleware.AuthRequired(jwtSecret))
	review.Post("/", reviewHandler.CreateReview)
	review.Put("/:id", reviewHandler.UpdateReview)
	review.Delete("/:id", reviewHandler.DeleteReview)
	review.Post("/:id/helpful", reviewHandler.MarkReviewHelpful)

	// Ticket routes
	ticket := api.Group("/tickets")
	ticket.Use(middleware.AuthRequired(jwtSecret))
	ticket.Post("/purchase", ticketHandler.PurchaseTicket)

	// Payment webhook (no auth required)
	api.Post("/payments/webhook", paymentHandler.PaymentWebhook)

	// Start server
	log.Fatal(app.Listen(":8080"))
}
