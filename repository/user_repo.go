
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
}
