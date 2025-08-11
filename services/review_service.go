package services

import (
	"errors"

	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
	"github.com/google/uuid"
)

type ReviewService interface {
	CreateReview(review *models.Review) error
	GetReviewByID(id uuid.UUID) (*models.Review, error)
	GetEventReviews(eventID uuid.UUID, page, limit int) ([]models.Review, error)
	GetHostReviews(hostID uuid.UUID, page, limit int) ([]models.Review, error)
	UpdateReview(id uuid.UUID, updates map[string]interface{}) error
	DeleteReview(id uuid.UUID) error
	GetEventRatingStats(eventID uuid.UUID) (map[string]interface{}, error)
	GetHostRatingStats(hostID uuid.UUID) (map[string]interface{}, error)
	MarkReviewHelpful(reviewID uuid.UUID) error
}

type reviewService struct {
	reviewRepo repository.ReviewRepository
}

func NewReviewService(reviewRepo repository.ReviewRepository) ReviewService {
	return &reviewService{
		reviewRepo: reviewRepo,
	}
}

func (s *reviewService) CreateReview(review *models.Review) error {
	// Validate rating
	if review.Rating < 1 || review.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	
	return s.reviewRepo.Create(review)
}

func (s *reviewService) GetReviewByID(id uuid.UUID) (*models.Review, error) {
	return s.reviewRepo.GetByID(id)
}

func (s *reviewService) GetEventReviews(eventID uuid.UUID, page, limit int) ([]models.Review, error) {
	offset := (page - 1) * limit
	return s.reviewRepo.GetByEventID(eventID, limit, offset)
}

func (s *reviewService) GetHostReviews(hostID uuid.UUID, page, limit int) ([]models.Review, error) {
	offset := (page - 1) * limit
	return s.reviewRepo.GetByHostID(hostID, limit, offset)
}

func (s *reviewService) UpdateReview(id uuid.UUID, updates map[string]interface{}) error {
	review, err := s.reviewRepo.GetByID(id)
	if err != nil {
		return err
	}
	
	// Update fields
	if rating, ok := updates["rating"]; ok {
		if r, ok := rating.(int); ok && r >= 1 && r <= 5 {
			review.Rating = r
		} else {
			return errors.New("invalid rating value")
		}
	}
	
	if comment, ok := updates["comment"]; ok {
		if c, ok := comment.(string); ok {
			review.Comment = c
		}
	}
	
	return s.reviewRepo.Update(review)
}

func (s *reviewService) DeleteReview(id uuid.UUID) error {
	return s.reviewRepo.Delete(id)
}

func (s *reviewService) GetEventRatingStats(eventID uuid.UUID) (map[string]interface{}, error) {
	distribution, average, err := s.reviewRepo.GetEventRatingStats(eventID)
	if err != nil {
		return nil, err
	}
	
	// Calculate total reviews
	totalReviews := 0
	for _, count := range distribution {
		totalReviews += count
	}
	
	return map[string]interface{}{
		"average_rating":      average,
		"total_reviews":       totalReviews,
		"rating_distribution": distribution,
	}, nil
}

func (s *reviewService) GetHostRatingStats(hostID uuid.UUID) (map[string]interface{}, error) {
	distribution, average, err := s.reviewRepo.GetHostRatingStats(hostID)
	if err != nil {
		return nil, err
	}
	
	// Calculate total reviews
	totalReviews := 0
	for _, count := range distribution {
		totalReviews += count
	}
	
	return map[string]interface{}{
		"average_rating":      average,
		"total_reviews":       totalReviews,
		"rating_distribution": distribution,
	}, nil
}

func (s *reviewService) MarkReviewHelpful(reviewID uuid.UUID) error {
	review, err := s.reviewRepo.GetByID(reviewID)
	if err != nil {
		return err
	}
	
	review.Helpful++
	return s.reviewRepo.Update(review)
}