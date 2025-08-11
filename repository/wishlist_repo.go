
package repository

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
)

type WishlistRepository interface {
	AddToWishlist(wishlist *models.Wishlist) error
	GetWishlistByUserID(userID uuid.UUID) ([]*models.Event, error)
}
