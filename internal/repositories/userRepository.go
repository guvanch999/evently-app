package repositories

import (
	"github.com/google/uuid"
	"github.com/guvanch999/evently-app/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db,
	}
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserById(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) UpdateUserProfile(user *models.User) error {
	return r.db.Model(user).Updates(models.User{
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		InstagramNickname: user.InstagramNickname,
		BirthDate:         user.BirthDate,
		UpdatedAt:         user.UpdatedAt,
	}).Error
}
