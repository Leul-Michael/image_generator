package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RequestStatus string

const (
	RequestStatusPending    RequestStatus = "pending"
	RequestStatusProcessing RequestStatus = "processing"
	RequestStatusCompleted  RequestStatus = "completed"
	RequestStatusFailed     RequestStatus = "failed"
)

type ImageGenerationRequest struct {
	Base
	UserID            uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	User              User            `gorm:"foreignKey:UserID" json:"user"`
	CategoryID        uuid.UUID       `gorm:"type:uuid;not null" json:"category_id"`
	Category          Category        `gorm:"foreignKey:CategoryID" json:"category"`
	Prompt            string          `gorm:"size:500;not null" json:"prompt"`
	ReferenceImageURL *string         `gorm:"size:500" json:"reference_image_url"`
	Status            RequestStatus   `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Error             *string         `gorm:"size:500" json:"error"`
	GeneratedImageID  *uuid.UUID      `gorm:"type:uuid" json:"generated_image_id"`
	GeneratedImage    *GeneratedImage `gorm:"foreignKey:GeneratedImageID" json:"generated_image"`
	CreditsRequired   int             `gorm:"not null" json:"credits_required"`
}

func (igr *ImageGenerationRequest) BeforeCreate(tx *gorm.DB) (err error) {
	igr.ID = uuid.New()
	return
}
