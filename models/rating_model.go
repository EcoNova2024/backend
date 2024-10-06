// backend/models/rating_model.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Rating represents the rating model
type Rating struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Score     float64   `gorm:"not null" json:"score"`
	CreatedAt time.Time `gorm:"default:current_timestamp" json:"created_at"`
}
type AddRating struct {
	ProductID string  `gorm:"type:uuid;not null" json:"product_id"`
	Score     float64 `gorm:"not null" json:"score"`
}

// BeforeCreate sets the UUID before creating a new record
func (r *Rating) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return
}
