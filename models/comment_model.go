// backend/models/Comment_model.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comment represents the Comment model in the database
type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Content   string    `gorm:"type:text;not null" json:"content"` // Changed to Content (text)
	CreatedAt time.Time `gorm:"default:current_timestamp" json:"created_at"`
}

// AddComment represents the structure to add a new comment to a product
type AddComment struct {
	ProductID string `gorm:"type:uuid;not null" json:"product_id"`
	Content   string `gorm:"type:text;not null" json:"content"` // Changed to Content
}

func (r *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New() // Automatically generate a new UUID for the Comment ID
	return
}

type CommentResponse struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Content   string    `gorm:"type:text;not null" json:"content"` // Changed to Content (text)
	CreatedAt time.Time `gorm:"default:current_timestamp" json:"created_at"`
}
