package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Leul-Michael/image-generation/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresUserRepo struct {
	DB *gorm.DB
}

type UserRepo interface {
	GetById(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByTelegramID(ctx context.Context, telegramID uint) (*model.User, error)
	CreateOrUpdateUser(ctx context.Context, telegramID uint, firstName, lastName string, username, photoURL *string) (*model.User, error)
	EmailExists(ctx context.Context, email string) int64
	ComparePassword(ctx context.Context, sub interface{}) (*Sub, error)
	Insert(ctx context.Context, user model.User) error
	Update(ctx context.Context, user model.User) error
	UpdateField(ctx context.Context, id uuid.UUID, field string, value interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
}

var ErrNotExist = errors.New("user not found")

func (pr *PostgresUserRepo) GetById(ctx context.Context, id uuid.UUID) (model.User, error) {
	var user model.User

	err := pr.DB.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Select("id, first_name, last_name, email, phone_number, last_login, is_deactivated").First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, ErrNotExist
		} else {
			return model.User{}, fmt.Errorf("failed to get record: %w", err)
		}
	}

	return user, nil
}

func (pr *PostgresUserRepo) EmailExists(ctx context.Context, email string) int64 {
	var count int64
	pr.DB.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count)
	return count
}

type Sub struct {
	Id            uuid.UUID `json:"id"`
	Password      string    `json:"password"`
	IsDeactivated bool      `json:"is_deactivated"`
}

func (pr *PostgresUserRepo) ComparePassword(ctx context.Context, email string) (*Sub, error) {
	var sub Sub
	err := pr.DB.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Select("id, password, is_deactivated").Scan(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (pr *PostgresUserRepo) Insert(ctx context.Context, user model.User) error {
	if err := pr.DB.WithContext(ctx).Model(&model.User{}).Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func (pr *PostgresUserRepo) Update(ctx context.Context, user model.User) error {
	if err := pr.DB.WithContext(ctx).Model(&model.User{}).Save(&user).Error; err != nil {
		return err
	}
	return nil
}

func (pr *PostgresUserRepo) UpdateField(ctx context.Context, id uuid.UUID, field string, value interface{}) error {
	if err := pr.DB.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update(field, value).
		Error; err != nil {
		return err
	}
	return nil
}

func (pr *PostgresUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := pr.DB.WithContext(ctx).Delete(&model.User{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (pr *PostgresUserRepo) GetByTelegramID(ctx context.Context, telegramID uint) (*model.User, error) {
	var user model.User
	err := pr.DB.WithContext(ctx).
		Where("telegram_id = ?", telegramID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotExist
		}
		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}
	return &user, nil
}

func (pr *PostgresUserRepo) CreateOrUpdateUser(ctx context.Context, telegramID uint, firstName, lastName string, username, photoURL *string) (*model.User, error) {
	tx := pr.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	var user model.User
	result := tx.Where("telegram_id = ?", telegramID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create new user
			user = model.User{
				FirstName:        firstName,
				LastName:         lastName,
				TelegramID:       telegramID,
				TelegramUsername: username,
				Image:            photoURL,
				Role:             model.RoleUser,
			}

			if err := tx.Create(&user).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create user: %w", err)
			}

			// Initialize user credits
			imageCredits := model.UserCredit{
				UserID:     user.ID,
				CreditType: model.CreditTypeImage,
				Balance:    0,
			}
			if err := tx.Create(&imageCredits).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create user credits: %w", err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("failed to check user existence: %w", result.Error)
		}
	} else {
		// Update existing user
		user.FirstName = firstName
		user.LastName = lastName
		user.TelegramUsername = username
		user.Image = photoURL
		now := time.Now()
		user.LastLogin = &now

		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &user, nil
}
