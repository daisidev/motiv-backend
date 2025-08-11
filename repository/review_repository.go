package repository

import (
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(review *models.Review) error
	GetByID(id uuid.UUID) (*models.Review, error)
	GetByEventID(eventID uuid.UUID, limit, offset int) ([]models.Review, error)
	GetByHostID(hostID uuid.UUID, limit, offset int) ([]models.Review, error)
	Update(review *models.Review) error
	Delete(id uuid.UUID) error
	GetEventRatingStats(eventID uuid.UUID) (map[int]int, float64, error)
	GetHostRatingStats(hostID uuid.UUID) (map[int]int, float64, error)
}

type reviewRepoPG struct {
	db *gorm.DB
}

func NewReviewRepoPG(db *gorm.DB) ReviewRepository {
	return &reviewRepoPG{db: db}
}

func (r *reviewRepoPG) Create(review *models.Review) error {
	return r.db.Create(review).Error
}

func (r *reviewRepoPG) GetByID(id uuid.UUID) (*models.Review, error) {
	var review models.Review
	err := r.db.Preload("User").Preload("Event").First(&review, "id = ?", id).Error
	return &review, err
}

func (r *reviewRepoPG) GetByEventID(eventID uuid.UUID, limit, offset int) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.Preload("User").Where("event_id = ?", eventID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepoPG) GetByHostID(hostID uuid.UUID, limit, offset int) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.Preload("User").Preload("Event").
		Joins("JOIN events ON reviews.event_id = events.id").
		Where("events.host_id = ?", hostID).
		Order("reviews.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepoPG) Update(review *models.Review) error {
	return r.db.Save(review).Error
}

func (r *reviewRepoPG) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Review{}, "id = ?", id).Error
}

func (r *reviewRepoPG) GetEventRatingStats(eventID uuid.UUID) (map[int]int, float64, error) {
	var results []struct {
		Rating int
		Count  int
	}
	
	err := r.db.Model(&models.Review{}).
		Select("rating, COUNT(*) as count").
		Where("event_id = ?", eventID).
		Group("rating").
		Find(&results).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	stats := make(map[int]int)
	totalRating := 0
	totalCount := 0
	
	for _, result := range results {
		stats[result.Rating] = result.Count
		totalRating += result.Rating * result.Count
		totalCount += result.Count
	}
	
	var averageRating float64
	if totalCount > 0 {
		averageRating = float64(totalRating) / float64(totalCount)
	}
	
	return stats, averageRating, nil
}

func (r *reviewRepoPG) GetHostRatingStats(hostID uuid.UUID) (map[int]int, float64, error) {
	var results []struct {
		Rating int
		Count  int
	}
	
	err := r.db.Model(&models.Review{}).
		Select("rating, COUNT(*) as count").
		Joins("JOIN events ON reviews.event_id = events.id").
		Where("events.host_id = ?", hostID).
		Group("rating").
		Find(&results).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	stats := make(map[int]int)
	totalRating := 0
	totalCount := 0
	
	for _, result := range results {
		stats[result.Rating] = result.Count
		totalRating += result.Rating * result.Count
		totalCount += result.Count
	}
	
	var averageRating float64
	if totalCount > 0 {
		averageRating = float64(totalRating) / float64(totalCount)
	}
	
	return stats, averageRating, nil
}