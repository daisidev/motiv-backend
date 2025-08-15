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
	Status      string `json:"status,omitempty"`

	// Date & Time
	StartDate string `json:"startDate" validate:"required"`
	StartTime string `json:"startTime" validate:"required"`
	EndTime   string `json:"endTime" validate:"required"`

	// Location
	Location    string                 `json:"location" validate:"required"`
	LocationData *LocationDataRequest `json:"locationData,omitempty"`

	// Tags
	Tags []string `json:"tags"`

	// Banner
	BannerImageURL string `json:"bannerImageURL"`

	// Tickets (only for ticketed events)
	TicketTypes []CreateTicketTypeRequest `json:"ticketTypes"`
}

// LocationDataRequest represents location data with coordinates
type LocationDataRequest struct {
	Address           string              `json:"address" validate:"required"`
	ManualDescription string              `json:"manualDescription,omitempty"`
	Coordinates       *CoordinatesRequest `json:"coordinates,omitempty"`
	PlaceID           *string             `json:"placeId,omitempty"`
}

// CoordinatesRequest represents latitude and longitude
type CoordinatesRequest struct {
	Lat float64 `json:"lat" validate:"required"`
	Lng float64 `json:"lng" validate:"required"`
}

// CreateTicketTypeRequest represents a ticket type in the creation request
type CreateTicketTypeRequest struct {
	Name          string  `json:"name" validate:"required"`
	Price         float64 `json:"price" validate:"min=0"`
	Description   string  `json:"description"`
	TotalQuantity int     `json:"totalQuantity" validate:"min=1"`
}

// EventResponse represents the response structure for events
type EventResponse struct {
	ID             uuid.UUID            `json:"id"`
	Title          string               `json:"title"`
	Description    string               `json:"description"`
	StartDate      time.Time            `json:"start_date"`
	StartTime      string               `json:"start_time"`
	EndTime        string               `json:"end_time"`
	Location       string               `json:"location"`
	Latitude       *float64             `json:"latitude,omitempty"`
	Longitude      *float64             `json:"longitude,omitempty"`
	PlaceID        *string              `json:"place_id,omitempty"`
	Tags           []string             `json:"tags"`
	BannerImageURL string               `json:"banner_image_url"`
	EventType      string               `json:"event_type"`
	HostID         uuid.UUID            `json:"host_id"`
	Host           UserResponse         `json:"host"`
	TicketTypes    []TicketTypeResponse `json:"ticket_types,omitempty"`
	Status         EventStatus          `json:"status"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
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

// GoogleAuthRequest represents the request payload for Firebase Google OAuth
type GoogleAuthRequest struct {
	IDToken string `json:"idToken" validate:"required"`
}

// GoogleUser represents the user data from Google OAuth
type GoogleUser struct {
	ID         string `json:"id" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Name       string `json:"name" validate:"required"`
	Picture    string `json:"picture,omitempty"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
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

// PaymentInitiationRequest represents the request to initiate payment
type PaymentInitiationRequest struct {
	EventID       string                `json:"eventId" validate:"required"`
	Email         string                `json:"email" validate:"required,email"`
	AttendeeData  AttendeeDataRequest   `json:"attendeeData" validate:"required"` // Primary attendee for payment
	Attendees     []AttendeeDataRequest `json:"attendees,omitempty"`              // All attendees (optional for backward compatibility)
	TicketDetails []TicketDetailRequest `json:"ticketDetails" validate:"required,min=1"`
}

// AttendeeDataRequest represents attendee information
type AttendeeDataRequest struct {
	FullName string `json:"fullName" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required"`
}

// MultipleAttendeesRequest represents multiple attendees for group bookings
type MultipleAttendeesRequest struct {
	Attendees []AttendeeDataRequest `json:"attendees" validate:"required,min=1"`
}

// TicketDetailRequest represents ticket purchase details
type TicketDetailRequest struct {
	TicketTypeID   string  `json:"ticketTypeId" validate:"required"`
	TicketTypeName string  `json:"ticketTypeName" validate:"required"`
	Quantity       int     `json:"quantity" validate:"required,min=1"`
	Price          float64 `json:"price" validate:"required,min=0"`
}

// PaymentInitiationResponse represents the response for payment initiation
type PaymentInitiationResponse struct {
	Reference   string `json:"reference"`
	Amount      int64  `json:"amount"` // Amount in kobo
	PaystackURL string `json:"paystackUrl"`
	PublicKey   string `json:"publicKey"`
	Email       string `json:"email"`
	Currency    string `json:"currency"`
}

// TicketResponse represents a purchased ticket
type TicketResponse struct {
	ID           uuid.UUID           `json:"id"`
	EventID      uuid.UUID           `json:"eventId"`
	Event        EventResponse       `json:"event"`
	TicketTypeID uuid.UUID           `json:"ticketTypeId"`
	TicketType   TicketTypeResponse  `json:"ticketType"`
	QRCode       string              `json:"qrCode"`
	AttendeeData AttendeeDataRequest `json:"attendeeData"`
	PurchasedAt  time.Time           `json:"purchasedAt"`
}

// ForgotPasswordRequest represents the request payload for forgot password
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents the request payload for resetting password
type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
}
