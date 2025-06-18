package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GeneratedImage struct {
	Base
	UserID            uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User              User      `gorm:"foreignKey:UserID" json:"user"`
	CategoryID        uuid.UUID `gorm:"type:uuid;not null" json:"category_id"`
	Category          Category  `gorm:"foreignKey:CategoryID" json:"category"`
	Prompt            string    `gorm:"size:500;not null" json:"prompt"`
	ImageURL          string    `gorm:"size:500;not null" json:"image_url"`
	ThumbnailURL      string    `gorm:"size:500" json:"thumbnail_url"`
	ReferenceImageURL *string   `gorm:"size:500" json:"reference_image_url"`
	Status            string    `gorm:"size:20;not null;default:'completed'" json:"status"`
	Error             *string   `gorm:"size:500" json:"error"`
	GenerationTime    int       `gorm:"not null" json:"generation_time"` // Time taken to generate in seconds
	CreditsUsed       int       `gorm:"not null" json:"credits_used"`
	IsPrivate         bool      `gorm:"default:true" json:"is_private"`

	// ChatGPT API specific fields
	ModelUsed        string `gorm:"size:50" json:"model_used"`         // e.g., "dall-e-3"
	PromptTokens     int    `gorm:"not null" json:"prompt_tokens"`     // Tokens used in the prompt
	CompletionTokens int    `gorm:"not null" json:"completion_tokens"` // Tokens in the completion
	TotalTokens      int    `gorm:"not null" json:"total_tokens"`      // Total tokens used
}

func (gi *GeneratedImage) BeforeCreate(tx *gorm.DB) (err error) {
	gi.ID = uuid.New()
	return
}
