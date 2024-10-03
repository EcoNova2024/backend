// backend/models/product.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the system
type Product struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	SubCategory string    `json:"sub_category"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}
type ProductResponse struct {
	ID           uuid.UUID    `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	User         User         `gorm:"type:user`
	transactions *Transaction `gorm:"type:transaction"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Price        float64      `json:"price"`
	SubCategory  string       `json:"sub_category"`
	Category     string       `json:"category"`
	CreatedAt    time.Time    `json:"created_at" gorm:"autoCreateTime"`
}
type ProductRequest struct {
	User         User         `gorm:"type:user`
	transactions *Transaction `gorm:"type:transaction"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Price        float64      `json:"price"`
	SubCategory  string       `json:"sub_category"`
	Category     string       `json:"category"`
}
type Transaction struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey;unique" json:"id"` // Primary key, unique identifier for each transaction
	ItemID      uuid.UUID         `gorm:"type:uuid;not null" json:"item_id"`     // Reference to the item involved in the transaction
	UserID      uuid.UUID         `gorm:"type:uuid;not null" json:"user_id"`     // Reference to the user performing the transaction
	Hidden      bool              `gorm:"hidden" json:"hidden"`
	Description string            `gorm:"type:text" json:"description"`                                         // Description of the transaction
	Action      TransactionAction `gorm:"type:enum('submitted', 'revitalized', 'sold');not null" json:"action"` // Action type of the transaction
	ImageURL    string            `gorm:"type:varchar(255)" json:"image_url"`                                   // URL of the transaction image
	CreatedAt   time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`                          // Transaction timestamp
}

// TransactionAction defines the possible actions for a transaction.
type TransactionAction string

const (
	Submitted   TransactionAction = "submitted"
	Revitalized TransactionAction = "revitalized"
	Sold        TransactionAction = "sold"
)
