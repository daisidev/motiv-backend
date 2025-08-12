package models

import (
	"time"
	"github.com/google/uuid"
)

// CreateEventRequest represents the request payload for creating an event
type CreateEventRequest struct {
	// Basic Details
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	EventType   string `json:"eventType" validate:"required,oneof=ticketed free"`
	
	// Date & Time
	StartDate string `json:"startDate" validate:"required"`
	StartTime string `json:"startTime" validate:"required"`
	EndTime   string `json:"endTime" validate:"required"`
	
	// Location
	Location string `json:"location" validate:"required"`
	
	// Tags
	Tags []string `json:"tags"`
	
	// Banner
	BannerImageURL string `json:"bannerImageURL"`
	
	// Tickets (only for ticketed events)
	TicketTypes []CreateTicketTypeRequest `json:"ticketTypes"`
}

// CreateTicketTypeRequest represents a ticket type in the creation request
type CreateTicketTypeRequest struct {
	Name            string  `json:"name" validate:"required"`
	Price           float64 `json:"price" validate:"min=0"`
	Description     string  `json:"description"`
	TotalQuantity   int     `json:"totalQuantity" validate:"min=1"`
}

// EventResponse represents the response structure for events
type EventResponse struct {
	ID             uuid.UUID           `json:"id"`
	Title          string              `json:"title"`
	Description    string              `json:"description"`
	StartDate      time.Time           `json:"start_date"`
	StartTime      string              `json:"start_time"`
	EndTime        string              `json:"end_time"`
	Location       string              `json:"location"`
	Tags           []string            `json:"tags"`
	BannerImageURL string              `json:"banner_image_url"`
	EventType      string              `json:"event_type"`
	HostID         uuid.UUID           `json:"host_id"`
	Host           UserResponse        `json:"host"`
	TicketTypes    []TicketTypeResponse `json:"ticket_types,omitempty"`
	Status         EventStatus         `json:"status"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// TicketTypeResponse represents the response structure for ticket types
type TicketTypeResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Price             float64   `json:"price"`
	Description       string    `json:"description"`
	TotalQuantity     int       `json:"total_quantity"`
	SoldQuantity      int       `json:"sold_quantity"`
	AvailableQuantity int       `json:"available_quantity"`
}

// SignupRequest represents the request payload for user registration
type SignupRequest struct {
	Name            string   `json:"name" validate:"required"`
	Username        string   `json:"username" validate:"required,min=3,max=30"`
	Email           string   `json:"email" validate:"required,email"`
	Password        string   `json:"password" validate:"required,min=6"`
	ConfirmPassword string   `json:"confirmPassword" validate:"required"`
	Role            UserRole `json:"role,omitempty"` // Optional role, defaults to guest
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse represents the response structure for users (to avoid password exposure)
type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar"`
	Role     string    `json:"role"`
}