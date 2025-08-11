
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

	// Create services
	userService := services.NewUserService(userRepo)
	eventService := services.NewEventService(eventRepo)
	ticketService := services.NewTicketService(ticketRepo)
	wishlistService := services.NewWishlistService(wishlistRepo)

	// Create handlers
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	authHandler := handlers.NewAuthHandler(userService, jwtSecret)
	userHandler := handlers.NewUserHandler(userService, wishlistService, ticketService)
	eventHandler := handlers.NewEventHandler(eventService)
	ticketHandler := handlers.NewTicketHandler(ticketService)

	// Create Fiber app
	app := fiber.New()

	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())

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

	// Event routes
	event := api.Group("/events")
	event.Get("/", eventHandler.GetAllEvents)
	event.Get("/:id", eventHandler.GetEventByID)

	// Host routes
	host := api.Group("/hosts")
	host.Use(middleware.AuthRequired(jwtSecret))
	host.Use(middleware.RoleRequired(models.HostRole, models.AdminRole, models.SuperhostRole))
	host.Get("/me/events", eventHandler.GetMyEvents)
	host.Post("/me/events", eventHandler.CreateEvent)
	host.Put("/me/events/:id", eventHandler.UpdateEvent)
	host.Delete("/me/events/:id", eventHandler.DeleteEvent)

	// Ticket routes
	ticket := api.Group("/tickets")
	ticket.Use(middleware.AuthRequired(jwtSecret))
	ticket.Post("/purchase", ticketHandler.PurchaseTicket)

	// Start server
	log.Fatal(app.Listen(":8080"))
}
