package services

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
)

type WishlistService interface {
	AddToWishlist(wishlist *models.Wishlist) error
	RemoveFromWishlist(userID, eventID uuid.UUID) error
	GetWishlistByUserID(userID uuid.UUID) ([]*models.Event, error)
	GetWishlistItemsByUserID(userID uuid.UUID) ([]*models.Wishlist, error)
	IsInWishlist(userID, eventID uuid.UUID) (bool, error)
}

type wishlistService struct {
	wishlistRepo repository.WishlistRepository
}

func NewWishlistService(wishlistRepo repository.WishlistRepository) WishlistService {
	return &wishlistService{wishlistRepo}
}

func (s *wishlistService) AddToWishlist(wishlist *models.Wishlist) error {
	// First check if the item already exists
	exists, err := s.wishlistRepo.IsInWishlist(wishlist.UserID, wishlist.EventID)
	if err != nil {
		return err
	}
	if exists {
		// Item already exists, return success (idempotent operation)
		return nil
	}
	return s.wishlistRepo.AddToWishlist(wishlist)
}

func (s *wishlistService) RemoveFromWishlist(userID, eventID uuid.UUID) error {
	return s.wishlistRepo.RemoveFromWishlist(userID, eventID)
}

func (s *wishlistService) GetWishlistByUserID(userID uuid.UUID) ([]*models.Event, error) {
	return s.wishlistRepo.GetWishlistByUserID(userID)
}

func (s *wishlistService) GetWishlistItemsByUserID(userID uuid.UUID) ([]*models.Wishlist, error) {
	return s.wishlistRepo.GetWishlistItemsByUserID(userID)
}

func (s *wishlistService) IsInWishlist(userID, eventID uuid.UUID) (bool, error) {
	return s.wishlistRepo.IsInWishlist(userID, eventID)
}
