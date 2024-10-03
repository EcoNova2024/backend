package models

import "github.com/google/uuid"

type Comment struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"not null"`
	ProductID uuid.UUID `gorm:"not null"`
	Content   string    `gorm:"not null"`
	Hidden    bool      `gorm:"default:false"`
	//Positivity int  `gorm:"default:50"`
}
type CommentResponse struct {
	ID        uuid.UUID `json:"id"`
	Content   string    `json:"content"`
	UserID    uuid.UUID `json:"user_id"`
	IsOwner   bool      `json:"is_owner"`
	ProductID uuid.UUID `json:"product_id"`
	//Positivity int  `gorm:"default:50"`
}
