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
	UserID     uuid.UUID  `gorm:"type:uuid;not null;constraint:OnDelete:CASCADE" json:"user_id"`
	User       User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	CreditType CreditType `gorm:"type:varchar(20);not null" json:"credit_type"`
	Credits    int        `gorm:"default:0" json:"credits"`
}

func (uc *UserCredit) BeforeCreate(tx *gorm.DB) (err error) {
	uc.ID = uuid.New()
	return
}

func (uc *UserCredit) UpdateBalance(tx *gorm.DB, amount int) error {
	uc.Credits += amount
	return tx.Save(uc).Error
}

func (uc *UserCredit) HasEnoughCredits(required int) bool {
	return uc.Credits >= required
}

func (uc *UserCredit) isValidCreditType() bool {
	return uc.CreditType == CreditTypeImage || uc.CreditType == CreditTypeVideo
}
