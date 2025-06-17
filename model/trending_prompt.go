package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrendingPrompt struct {
	Base
	Prompt     string    `gorm:"size:500;not null" json:"prompt"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null" json:"category_id"`
	Category   Category  `gorm:"foreignKey:CategoryID" json:"category"`
	UseCount   int       `gorm:"default:0" json:"use_count"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	LastUsedAt time.Time `json:"last_used_at"`
}

func (tp *TrendingPrompt) BeforeCreate(tx *gorm.DB) (err error) {
	tp.ID = uuid.New()
	return
}
