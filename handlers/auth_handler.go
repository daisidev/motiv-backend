
package handlers

import (
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/services"
)

// AuthHandler handles authentication-related requests

type AuthHandler struct {
	userService services.UserService
	jwtSecret   []byte
}

func NewAuthHandler(userService services.UserService, jwtSecret []byte) *AuthHandler {
	return &AuthHandler{userService, jwtSecret}
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) == 0 {
		return false
	}
	
	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Helper function to validate username format
func isValidUsername(username string) bool {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	
	// Username can only contain letters, numbers, and underscores
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return usernameRegex.MatchString(username)
}

// Signup handles user registration
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var signupReq models.SignupRequest
	if err := c.BodyParser(&signupReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate required fields
	if signupReq.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}
	if signupReq.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username is required"})
	}
	if signupReq.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email is required"})
	}
	if signupReq.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password is required"})
	}
	if signupReq.ConfirmPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password confirmation is required"})
	}

	// Validate field lengths and formats
	if len(strings.TrimSpace(signupReq.Name)) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name must be at least 2 characters long"})
	}
	if !isValidUsername(signupReq.Username) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username must be 3-30 characters and contain only letters, numbers, and underscores"})
	}
	if len(signupReq.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password must be at least 6 characters long"})
	}

	// Validate email format
	if !isValidEmail(signupReq.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email format"})
	}

	// Validate password confirmation
	if signupReq.Password != signupReq.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Passwords do not match"})
	}

	// Check if email already exists
	if existingUser, _ := h.userService.GetUserByEmail(signupReq.Email); existingUser != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email already exists"})
	}

	// Check if username already exists
	if existingUser, _ := h.userService.GetUserByUsername(signupReq.Username); existingUser != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username already exists"})
	}

	// Validate and set role
	role := signupReq.Role
	if role == "" {
		role = models.GuestRole // Default role if not specified
	}
	
	// Validate role is valid
	validRoles := []models.UserRole{models.GuestRole, models.HostRole, models.AdminRole, models.SuperhostRole}
	isValidRole := false
	for _, validRole := range validRoles {
		if role == validRole {
			isValidRole = true
			break
		}
	}
	if !isValidRole {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role specified"})
	}
	
	// Create user model from request (trim whitespace)
	newUser := models.User{
		Name:     strings.TrimSpace(signupReq.Name),
		Username: strings.TrimSpace(signupReq.Username),
		Email:    strings.ToLower(strings.TrimSpace(signupReq.Email)), // Normalize email to lowercase
		Password: signupReq.Password, // Don't trim password as spaces might be intentional
		Role:     role,
	}

	if err := h.userService.CreateUser(&newUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	// Create token for automatic login after signup
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = newUser.ID
	claims["role"] = newUser.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token
	t, err := token.SignedString(h.jwtSecret)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Return user response with token (same as login)
	userResponse := models.UserResponse{
		ID:       newUser.ID,
		Name:     newUser.Name,
		Username: newUser.Username,
		Email:    newUser.Email,
		Avatar:   newUser.Avatar,
		Role:     string(newUser.Role),
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token": t,
		"user":  userResponse,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var loginReq models.LoginRequest
	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate required fields
	if loginReq.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email is required"})
	}
	if loginReq.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password is required"})
	}

	// Validate email format
	if !isValidEmail(loginReq.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email format"})
	}

	user, err := h.userService.LoginUser(loginReq.Email, loginReq.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token
	t, err := token.SignedString(h.jwtSecret)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Return user data with token
	userResponse := models.UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Role:     string(user.Role),
	}

	return c.JSON(fiber.Map{
		"token": t,
		"user":  userResponse,
	})
}
