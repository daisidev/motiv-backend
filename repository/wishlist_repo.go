package repository

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
)

type WishlistRepository interface {
	AddToWishlist(wishlist *models.Wishlist) error
	RemoveFromWishlist(userID, eventID uuid.UUID) error
	GetWishlistByUserID(userID uuid.UUID) ([]*models.Event, error)
	GetWishlistItemsByUserID(userID uuid.UUID) ([]*models.Wishlist, error)
	IsInWishlist(userID, eventID uuid.UUID) (bool, error)
}
