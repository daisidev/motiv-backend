package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
)

type AdminHandler struct {
	userService    services.UserService
	paymentService services.PaymentService
	eventService   services.EventService
	ticketService  services.TicketService
}

func NewAdminHandler(userService services.UserService, paymentService services.PaymentService, eventService services.EventService, ticketService services.TicketService) *AdminHandler {
	return &AdminHandler{
		userService:    userService,
		paymentService: paymentService,
		eventService:   eventService,
		ticketService:  ticketService,
	}
}

// GetDashboardStats returns overall platform statistics
func (h *AdminHandler) GetDashboardStats(c *fiber.Ctx) error {
	// Verify admin role
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	role := claims["role"].(string)

	if role != string(models.AdminRole) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin access required"})
	}

	stats, err := h.userService.GetPlatformStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get platform stats"})
	}

	return c.JSON(stats)
}

// GetAllUsers returns paginated list of all users
func (h *AdminHandler) GetAllUsers(c *fiber.Ctx) error {
	// Verify admin role
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	role := claims["role"].(string)

	if role != string(models.AdminRole) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin access required"})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	roleFilter := c.Query("role", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	users, total, err := h.userService.GetAllUsersWithFilters(limit, offset, search, roleFilter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get users"})
	}

	return c.JSON(fiber.Map{
		"data":    users,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"hasMore": int64(offset+limit) < total,
	})
}

// GetAllTransactions returns paginated list of all transactions
func (h *AdminHandler) GetAllTransactions(c *fiber.Ctx) error {
	// Verify admin role
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	role := claims["role"].(string)

	if role != string(models.AdminRole) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin access required"})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	status := c.Query("status", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	transactions, total, err := h.paymentService.GetAllTransactionsWithFilters(limit, offset, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get transactions"})
	}

	return c.JSON(fiber.Map{
		"data":    transactions,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"hasMore": int64(offset+limit) < total,
	})
}

// GetUserDetails returns detailed information about a specific user
func (h *AdminHandler) GetUserDetails(c *fiber.Ctx) error {
	// Verify admin role
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	role := claims["role"].(string)

	if role != string(models.AdminRole) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin access required"})
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	userDetails, err := h.userService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Get user's tickets
	tickets, _ := h.ticketService.GetTicketsByUserID(userID)

	// Get user's events if they're a host
	var events []*models.Event
	if userDetails.Role == models.HostRole || userDetails.Role == models.AdminRole {
		events, _ = h.eventService.GetEventsByHostID(userID)
	}

	return c.JSON(fiber.Map{
		"user":    userDetails,
		"tickets": tickets,
		"events":  events,
	})
}

// UpdateUserRole allows admin to update user roles
func (h *AdminHandler) UpdateUserRole(c *fiber.Ctx) error {
	// Verify admin role
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	role := claims["role"].(string)

	if role != string(models.AdminRole) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin access required"})
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var request struct {
		Role string `json:"role"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate role
	validRoles := []models.UserRole{models.GuestRole, models.HostRole, models.AdminRole, models.SuperhostRole}
	isValidRole := false
	for _, validRole := range validRoles {
		if request.Role == string(validRole) {
			isValidRole = true
			break
		}
	}
	if !isValidRole {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role"})
	}

	err = h.userService.UpdateUserRole(userID, models.UserRole(request.Role))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user role"})
	}

	return c.JSON(fiber.Map{"message": "User role updated successfully"})
}
