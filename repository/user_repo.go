
package repository

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	UpdateUser(user *models.User) error
	CreatePasswordResetToken(token *models.PasswordResetToken) error
	GetPasswordResetToken(token string) (*models.PasswordResetToken, error)
	MarkPasswordResetTokenAsUsed(tokenID uuid.UUID) error
	UpdateUserPassword(userID uuid.UUID, hashedPassword string) error
}
