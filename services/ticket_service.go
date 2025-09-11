package services

import (
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
	"github.com/skip2/go-qrcode"
)

type TicketService interface {
	PurchaseTicket(ticket *models.Ticket) error
	CreateTicketWithQR(ticket *models.Ticket) error
	UpdateTicket(ticket *models.Ticket) error
	GetTicketsByUserID(userID uuid.UUID) ([]*models.Ticket, error)
	GetTicketByID(id uuid.UUID) (*models.Ticket, error)
	DecodeQRData(qrData string) (ticketID, eventID, userID uuid.UUID, err error)

	// Ticket Type methods
	CreateTicketType(ticketType *models.TicketType) error
	GetTicketTypesByEventID(eventID uuid.UUID) ([]*models.TicketType, error)
	GetTicketTypeByID(ticketTypeID uuid.UUID) (*models.TicketType, error)
	UpdateSoldQuantity(ticketTypeID uuid.UUID, quantity int) error
}

type ticketService struct {
	ticketRepo   repository.TicketRepository
	attendeeRepo repository.AttendeeRepository
}

func NewTicketService(ticketRepo repository.TicketRepository, attendeeRepo repository.AttendeeRepository) TicketService {
	return &ticketService{
		ticketRepo:   ticketRepo,
		attendeeRepo: attendeeRepo,
	}
}

func (s *ticketService) PurchaseTicket(ticket *models.Ticket) error {
	// In a real application, you would have more logic here, e.g., payment processing, QR code generation, etc.
	return s.ticketRepo.CreateTicket(ticket)
}

func (s *ticketService) CreateTicketWithQR(ticket *models.Ticket) error {
	// Validate that EventID and UserID are not nil
	if ticket.EventID == uuid.Nil {
		return fmt.Errorf("event ID cannot be nil")
	}
	if ticket.UserID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}
	if ticket.TicketTypeID == uuid.Nil {
		return fmt.Errorf("ticket type ID cannot be nil")
	}

	// First create the ticket to get the ID
	err := s.ticketRepo.CreateTicket(ticket)
	if err != nil {
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	// Generate QR code data with the actual ticket ID
	qrData := fmt.Sprintf("MOTIV-TICKET:%s:%s:%s", ticket.ID.String(), ticket.EventID.String(), ticket.UserID.String())
	
	// Generate QR code image as base64 string
	qrCodeBytes, err := qrcode.Encode(qrData, qrcode.Medium, 256)
	if err != nil {
		return fmt.Errorf("failed to generate QR code image: %w", err)
	}
	
	// Convert to base64 for storage/transmission
	qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCodeBytes)
	ticket.QRCode = qrCodeBase64

	// Update the ticket with the QR code
	err = s.ticketRepo.UpdateTicket(ticket)
	if err != nil {
		return fmt.Errorf("failed to update ticket with QR code: %w", err)
	}

	// Create corresponding attendee record
	attendee := &models.Attendee{
		EventID:  ticket.EventID,
		UserID:   ticket.UserID,
		TicketID: ticket.ID,
		Status:   models.AttendeeActive,
	}

	err = s.attendeeRepo.Create(attendee)
	if err != nil {
		return fmt.Errorf("failed to create attendee record: %w", err)
	}

	return nil
}

func (s *ticketService) GetTicketTypeByID(ticketTypeID uuid.UUID) (*models.TicketType, error) {
	return s.ticketRepo.GetTicketTypeByID(ticketTypeID)
}

func (s *ticketService) UpdateSoldQuantity(ticketTypeID uuid.UUID, quantity int) error {
	return s.ticketRepo.UpdateSoldQuantity(ticketTypeID, quantity)
}

func (s *ticketService) GetTicketsByUserID(userID uuid.UUID) ([]*models.Ticket, error) {
	return s.ticketRepo.GetTicketsByUserID(userID)
}

func (s *ticketService) UpdateTicket(ticket *models.Ticket) error {
	return s.ticketRepo.UpdateTicket(ticket)
}

func (s *ticketService) GetTicketByID(id uuid.UUID) (*models.Ticket, error) {
	return s.ticketRepo.GetTicketByID(id)
}

// Ticket Type methods
func (s *ticketService) CreateTicketType(ticketType *models.TicketType) error {
	return s.ticketRepo.CreateTicketType(ticketType)
}

func (s *ticketService) GetTicketTypesByEventID(eventID uuid.UUID) ([]*models.TicketType, error) {
	return s.ticketRepo.GetTicketTypesByEventID(eventID)
}

// DecodeQRData decodes QR code data to extract ticket information
func (s *ticketService) DecodeQRData(qrData string) (ticketID, eventID, userID uuid.UUID, err error) {
	// QR data format: "MOTIV-TICKET:{ticketID}:{eventID}:{userID}"
	parts := fmt.Sprintf("%s", qrData)
	
	// Split by colon
	segments := make([]string, 0)
	current := ""
	for _, char := range parts {
		if char == ':' {
			segments = append(segments, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	segments = append(segments, current) // Add the last segment
	
	if len(segments) != 4 || segments[0] != "MOTIV-TICKET" {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("invalid QR code format")
	}
	
	ticketID, err = uuid.Parse(segments[1])
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("invalid ticket ID in QR code")
	}
	
	eventID, err = uuid.Parse(segments[2])
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("invalid event ID in QR code")
	}
	
	userID, err = uuid.Parse(segments[3])
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("invalid user ID in QR code")
	}
	
	return ticketID, eventID, userID, nil
}

// RefreshTicketEventData refreshes event data for tickets that might have missing relationships
func (s *ticketService) RefreshTicketEventData(userID uuid.UUID) error {
	tickets, err := s.ticketRepo.GetTicketsByUserID(userID)
	if err != nil {
		return err
	}

	for _, ticket := range tickets {
		if ticket.Event.ID == uuid.Nil || ticket.Event.Title == "" {
			// Force reload the ticket with proper relationships
			refreshedTicket, err := s.ticketRepo.GetTicketByID(ticket.ID)
			if err != nil {
				continue
			}
			// Update the ticket in place if needed
			if refreshedTicket.Event.ID != uuid.Nil {
				ticket.Event = refreshedTicket.Event
			}
		}
	}

	return nil
}
