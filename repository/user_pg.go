
package repository

import (
	"github.com/google/uuid"
	"github.com/hidenkeys/motiv-backend/models"
	"gorm.io/gorm"
)

type userRepoPG struct {
	db *gorm.DB
}

func NewUserRepoPG(db *gorm.DB) UserRepository {
	return &userRepoPG{db}
}

func (r *userRepoPG) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepoPG) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepoPG) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepoPG) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepoPG) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}
