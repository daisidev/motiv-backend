package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
)

type AttendeeService interface {
	CreateAttendee(attendee *models.Attendee) error
	GetAttendeeByID(id uuid.UUID) (*models.Attendee, error)
	GetEventAttendees(eventID uuid.UUID, limit, offset int) ([]models.Attendee, error)
	GetHostAttendees(hostID uuid.UUID, limit, offset int) ([]models.Attendee, error)
	CheckInByQRCode(qrCode string, eventID, checkedInBy uuid.UUID) (*CheckInResult, error)
	GetEventAttendeeStats(eventID uuid.UUID) (map[string]int64, error)
	GetHostAttendeeStats(hostID uuid.UUID) (map[string]int64, error)
}

type CheckInResult struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	Attendee  *AttendeeResponse   `json:"attendee,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
}

type AttendeeResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Event       string    `json:"event"`
	TicketType  string    `json:"ticketType"`
	QRCode      string    `json:"qrCode"`
	Status      string    `json:"status"`
	CheckedInAt *time.Time `json:"checkedInAt,omitempty"`
}

type attendeeService struct {
	attendeeRepo repository.AttendeeRepository
	ticketRepo   repository.TicketRepository
}

func NewAttendeeService(attendeeRepo repository.AttendeeRepository, ticketRepo repository.TicketRepository) AttendeeService {
	return &attendeeService{
		attendeeRepo: attendeeRepo,
		ticketRepo:   ticketRepo,
	}
}

func (s *attendeeService) CreateAttendee(attendee *models.Attendee) error {
	return s.attendeeRepo.Create(attendee)
}

func (s *attendeeService) GetAttendeeByID(id uuid.UUID) (*models.Attendee, error) {
	return s.attendeeRepo.GetByID(id)
}

func (s *attendeeService) GetEventAttendees(eventID uuid.UUID, limit, offset int) ([]models.Attendee, error) {
	return s.attendeeRepo.GetByEventID(eventID, limit, offset)
}

func (s *attendeeService) GetHostAttendees(hostID uuid.UUID, limit, offset int) ([]models.Attendee, error) {
	return s.attendeeRepo.GetByHostID(hostID, limit, offset)
}

func (s *attendeeService) CheckInByQRCode(qrCode string, eventID, checkedInBy uuid.UUID) (*CheckInResult, error) {
	// Find ticket by QR code
	ticket, err := s.ticketRepo.GetByQRCode(qrCode)
	if err != nil {
		return &CheckInResult{
			Success:   false,
			Message:   "Invalid QR Code - Ticket not found",
			Timestamp: time.Now(),
		}, nil
	}

	// Verify ticket belongs to the correct event
	if ticket.EventID != eventID {
		return &CheckInResult{
			Success:   false,
			Message:   "Wrong event - This ticket is for a different event",
			Timestamp: time.Now(),
		}, nil
	}

	// Find attendee by ticket ID
	attendees, err := s.attendeeRepo.GetByEventID(eventID, 1000, 0) // Get all attendees for the event
	if err != nil {
		return nil, err
	}

	var targetAttendee *models.Attendee
	for _, attendee := range attendees {
		if attendee.TicketID == ticket.ID {
			targetAttendee = &attendee
			break
		}
	}

	if targetAttendee == nil {
		return &CheckInResult{
			Success:   false,
			Message:   "Attendee not found for this ticket",
			Timestamp: time.Now(),
		}, nil
	}

	// Check if already checked in
	if targetAttendee.Status == models.AttendeeCheckedIn {
		return &CheckInResult{
			Success:   false,
			Message:   "Already checked in",
			Attendee: &AttendeeResponse{
				ID:          targetAttendee.ID,
				Name:        targetAttendee.User.Name,
				Email:       targetAttendee.User.Email,
				Event:       targetAttendee.Event.Title,
				TicketType:  ticket.TicketType.Name,
				QRCode:      ticket.QRCode,
				Status:      string(targetAttendee.Status),
				CheckedInAt: targetAttendee.CheckedInAt,
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Check if cancelled
	if targetAttendee.Status == models.AttendeeCancelled {
		return &CheckInResult{
			Success:   false,
			Message:   "Ticket has been cancelled",
			Attendee: &AttendeeResponse{
				ID:         targetAttendee.ID,
				Name:       targetAttendee.User.Name,
				Email:      targetAttendee.User.Email,
				Event:      targetAttendee.Event.Title,
				TicketType: ticket.TicketType.Name,
				QRCode:     ticket.QRCode,
				Status:     string(targetAttendee.Status),
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Check in the attendee
	err = s.attendeeRepo.CheckInAttendee(targetAttendee.ID, checkedInBy)
	if err != nil {
		return nil, err
	}

	// Get updated attendee
	updatedAttendee, err := s.attendeeRepo.GetByID(targetAttendee.ID)
	if err != nil {
		return nil, err
	}

	return &CheckInResult{
		Success: true,
		Message: "Successfully checked in!",
		Attendee: &AttendeeResponse{
			ID:          updatedAttendee.ID,
			Name:        updatedAttendee.User.Name,
			Email:       updatedAttendee.User.Email,
			Event:       updatedAttendee.Event.Title,
			TicketType:  ticket.TicketType.Name,
			QRCode:      ticket.QRCode,
			Status:      string(updatedAttendee.Status),
			CheckedInAt: updatedAttendee.CheckedInAt,
		},
		Timestamp: time.Now(),
	}, nil
}

func (s *attendeeService) GetEventAttendeeStats(eventID uuid.UUID) (map[string]int64, error) {
	return s.attendeeRepo.GetEventAttendeeStats(eventID)
}

func (s *attendeeService) GetHostAttendeeStats(hostID uuid.UUID) (map[string]int64, error) {
	return s.attendeeRepo.GetHostAttendeeStats(hostID)
}