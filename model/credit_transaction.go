package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypePurchase TransactionType = "purchase"
	TransactionTypeUsage    TransactionType = "usage"
	TransactionTypeRefund   TransactionType = "refund"
)

type Transaction struct {
	Base
	UserID           uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	User             User            `gorm:"foreignKey:UserID" json:"user"`
	CreditType       CreditType      `gorm:"type:varchar(20);not null" json:"credit_type"`
	Amount           int             `gorm:"not null" json:"amount"` // Can be positive (purchase) or negative (usage)
	Type             TransactionType `gorm:"type:varchar(20);not null" json:"type"`
	Description      string          `gorm:"size:500" json:"description"`
	BalanceAfter     int             `gorm:"not null" json:"balance_after"`
	ReferenceID      *string         `gorm:"size:100" json:"reference_id"` // For external payment references
	PaymentProvider  *string         `gorm:"size:50" json:"payment_provider"`
	GeneratedImageID *uuid.UUID      `gorm:"type:uuid" json:"generated_image_id"` // Link to the generated image if this is a usage transaction
	GeneratedImage   *GeneratedImage `gorm:"foreignKey:GeneratedImageID" json:"generated_image"`
}

func (ct *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	ct.ID = uuid.New()
	return
}
