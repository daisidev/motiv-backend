
package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"github.com/hidenkeys/motiv-backend/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(user *models.User) error
	LoginUser(email, password string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	UpdateUser(user *models.User) error
	CreatePasswordResetToken(userID uuid.UUID, token string, expiresAt time.Time) error
	GetPasswordResetToken(token string) (*models.PasswordResetToken, error)
	MarkPasswordResetTokenAsUsed(tokenID uuid.UUID) error
	UpdateUserPassword(userID uuid.UUID, newPassword string) error
	
	// Admin methods
	GetPlatformStats() (map[string]interface{}, error)
	GetAllUsersWithFilters(limit, offset int, search, roleFilter string) ([]*models.User, int64, error)
	UpdateUserRole(userID uuid.UUID, role models.UserRole) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo}
}

func (s *userService) CreateUser(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return s.userRepo.CreateUser(user)
}

// Export error variables so they can be used by handlers
var ErrUserNotFound = errors.New("user not found")
var ErrInvalidPassword = errors.New("invalid password")

func (s *userService) LoginUser(email, password string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.GetUserByEmail(email)
}

func (s *userService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepo.GetUserByUsername(username)
}

func (s *userService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetUserByID(id)
}

func (s *userService) UpdateUser(user *models.User) error {
	return s.userRepo.UpdateUser(user)
}

func (s *userService) CreatePasswordResetToken(userID uuid.UUID, token string, expiresAt time.Time) error {
	resetToken := &models.PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
	}
	return s.userRepo.CreatePasswordResetToken(resetToken)
}

func (s *userService) GetPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	return s.userRepo.GetPasswordResetToken(token)
}

func (s *userService) MarkPasswordResetTokenAsUsed(tokenID uuid.UUID) error {
	return s.userRepo.MarkPasswordResetTokenAsUsed(tokenID)
}

func (s *userService) UpdateUserPassword(userID uuid.UUID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateUserPassword(userID, string(hashedPassword))
}

// Admin methods
func (s *userService) GetPlatformStats() (map[string]interface{}, error) {
	return s.userRepo.GetPlatformStats()
}

func (s *userService) GetAllUsersWithFilters(limit, offset int, search, roleFilter string) ([]*models.User, int64, error) {
	return s.userRepo.GetAllUsersWithFilters(limit, offset, search, roleFilter)
}

func (s *userService) UpdateUserRole(userID uuid.UUID, role models.UserRole) error {
	return s.userRepo.UpdateUserRole(userID, role)
}
