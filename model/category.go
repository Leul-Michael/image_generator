package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	Base
	Name        string `gorm:"size:100;not null;unique" json:"name"`
	Description string `gorm:"size:500" json:"description"`
	Emoji       string `gorm:"size:50" json:"emoji"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}
