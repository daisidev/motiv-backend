
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
