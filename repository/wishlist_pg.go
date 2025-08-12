package repository

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"gorm.io/gorm"
)

type wishlistRepoPG struct {
	db *gorm.DB
}

func NewWishlistRepoPG(db *gorm.DB) WishlistRepository {
	return &wishlistRepoPG{db}
}

func (r *wishlistRepoPG) AddToWishlist(wishlist *models.Wishlist) error {
	return r.db.Create(wishlist).Error
}

func (r *wishlistRepoPG) RemoveFromWishlist(userID, eventID uuid.UUID) error {
	return r.db.Where("user_id = ? AND event_id = ?", userID, eventID).Delete(&models.Wishlist{}).Error
}

func (r *wishlistRepoPG) GetWishlistByUserID(userID uuid.UUID) ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.Table("events").
		Joins("join wishlists on wishlists.event_id = events.id").
		Where("wishlists.user_id = ?", userID).
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *wishlistRepoPG) GetWishlistItemsByUserID(userID uuid.UUID) ([]*models.Wishlist, error) {
	var wishlistItems []*models.Wishlist
	err := r.db.Where("user_id = ?", userID).
		Preload("Event").
		Preload("User").
		Find(&wishlistItems).Error
	if err != nil {
		return nil, err
	}
	return wishlistItems, nil
}

func (r *wishlistRepoPG) IsInWishlist(userID, eventID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Wishlist{}).
		Where("user_id = ? AND event_id = ?", userID, eventID).
		Count(&count).Error
	return count > 0, err
}
