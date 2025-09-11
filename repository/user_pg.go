
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

// Admin methods
func (r *userRepoPG) GetPlatformStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total users
	var totalUsers int64
	r.db.Model(&models.User{}).Count(&totalUsers)
	stats["total_users"] = totalUsers

	// Users by role
	var guestCount, hostCount, adminCount int64
	r.db.Model(&models.User{}).Where("role = ?", models.GuestRole).Count(&guestCount)
	r.db.Model(&models.User{}).Where("role = ?", models.HostRole).Count(&hostCount)
	r.db.Model(&models.User{}).Where("role = ?", models.AdminRole).Count(&adminCount)
	
	stats["guest_users"] = guestCount
	stats["host_users"] = hostCount
	stats["admin_users"] = adminCount

	// Total events
	var totalEvents int64
	r.db.Table("events").Count(&totalEvents)
	stats["total_events"] = totalEvents

	// Total tickets
	var totalTickets int64
	r.db.Table("tickets").Count(&totalTickets)
	stats["total_tickets"] = totalTickets

	// Newsletter subscribers
	var newsletterSubscribers int64
	r.db.Model(&models.User{}).Where("newsletter_subscribed = ?", true).Count(&newsletterSubscribers)
	stats["newsletter_subscribers"] = newsletterSubscribers

	return stats, nil
}

func (r *userRepoPG) GetAllUsersWithFilters(limit, offset int, search, roleFilter string) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply search filter
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ? OR username ILIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Apply role filter
	if roleFilter != "" {
		query = query.Where("role = ?", roleFilter)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepoPG) UpdateUserRole(userID uuid.UUID, role models.UserRole) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}
