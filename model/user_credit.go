package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreditType string

const (
	CreditTypeImage CreditType = "image"
	CreditTypeVideo CreditType = "video"
)

type UserCredit struct {
	Base
	UserID     uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User       User       `gorm:"foreignKey:UserID" json:"user"`
	CreditType CreditType `gorm:"type:varchar(20);not null" json:"credit_type"`
	Balance    int        `gorm:"default:0" json:"balance"`
}

func (uc *UserCredit) BeforeCreate(tx *gorm.DB) (err error) {
	uc.ID = uuid.New()
	return
}

// UpdateBalance updates the credit balance
func (uc *UserCredit) UpdateBalance(tx *gorm.DB, amount int) error {
	uc.Balance += amount
	return tx.Save(uc).Error
}

// HasEnoughCredits checks if the user has enough credits of this type
func (uc *UserCredit) HasEnoughCredits(required int) bool {
	return uc.Balance >= required
}

func (uc *UserCredit) isValidCreditType() bool {
	return uc.CreditType == CreditTypeImage || uc.CreditType == CreditTypeVideo
}
