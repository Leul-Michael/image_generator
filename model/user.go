package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleUser       Role = "user"
)

type User struct {
	Base
	FirstName        string       `gorm:"size:100;not null" json:"first_name"`
	LastName         string       `gorm:"size:100;not null" json:"last_name"`
	Email            *string      `gorm:"size:100;unique" json:"email"`
	Image            *string      `gorm:"size:255" json:"image"`
	TelegramID       uint         `gorm:"unique;not null" json:"telegram_id"`
	TelegramUsername *string      `gorm:"size:100;unique;not null" json:"telegram_username"`
	Password         *string      `json:"-"`
	PhoneNumber      *string      `json:"phone_number"`
	LastLogin        *time.Time   `json:"last_login"`
	Role             Role         `gorm:"default:'user'" json:"role"`
	IsDeactivated    bool         `gorm:"default:false" json:"is_deactivated"`
	UserCredits      []UserCredit `gorm:"foreignKey:UserID" json:"user_credits"`
	Lang             string       `gorm:"default:'en'" json:"lang"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New()
	return
}

func (u *User) GetCreditBalance(tx *gorm.DB, creditType CreditType) (int, error) {
	var userCredit UserCredit
	err := tx.Where("user_id = ? AND credit_type = ?", u.ID, creditType).First(&userCredit).Error
	if err != nil {
		return 0, err
	}
	return userCredit.Credits, nil
}

func (u *User) GetCurrentBalance(tx *gorm.DB, creditType CreditType) (int, error) {
	var balance int
	err := tx.Model(&Transaction{}).
		Where("user_id = ? AND credit_type = ?", u.ID, creditType).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&balance).Error
	return balance, err
}

func (u *User) HasEnoughCredits(tx *gorm.DB, creditType CreditType, required int) (bool, error) {
	balance, err := u.GetCreditBalance(tx, creditType)
	if err != nil {
		return false, err
	}
	return balance >= required, nil
}

func (u *User) GetTokenUsage(tx *gorm.DB, creditType CreditType) (*UserCredit, error) {
	var userCredit UserCredit
	err := tx.Where("user_id = ? AND credit_type = ?", u.ID, creditType).First(&userCredit).Error
	if err != nil {
		return nil, err
	}
	return &userCredit, nil
}
