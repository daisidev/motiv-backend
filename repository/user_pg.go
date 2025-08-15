
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

func (r *userRepoPG) CreatePasswordResetToken(token *models.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *userRepoPG) GetPasswordResetToken(token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	err := r.db.Preload("User").Where("token = ? AND used = ? AND expires_at > NOW()", token, false).First(&resetToken).Error
	if err != nil {
		return nil, err
	}
	return &resetToken, nil
}

func (r *userRepoPG) MarkPasswordResetTokenAsUsed(tokenID uuid.UUID) error {
	return r.db.Model(&models.PasswordResetToken{}).Where("id = ?", tokenID).Update("used", true).Error
}

func (r *userRepoPG) UpdateUserPassword(userID uuid.UUID, hashedPassword string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}
