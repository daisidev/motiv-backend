
package services

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
)

type WishlistService interface {
	AddToWishlist(wishlist *models.Wishlist) error
	GetWishlistByUserID(userID uuid.UUID) ([]*models.Event, error)
}

type wishlistService struct {
	wishlistRepo repository.WishlistRepository
}

func NewWishlistService(wishlistRepo repository.WishlistRepository) WishlistService {
	return &wishlistService{wishlistRepo}
}

func (s *wishlistService) AddToWishlist(wishlist *models.Wishlist) error {
	return s.wishlistRepo.AddToWishlist(wishlist)
}

func (s *wishlistService) GetWishlistByUserID(userID uuid.UUID) ([]*models.Event, error) {
	return s.wishlistRepo.GetWishlistByUserID(userID)
}
