package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductStatus defines the possible statuses for a product.
type ProductStatus string

const (
	StatusAvailable ProductStatus = "available"
	StatusRestored  ProductStatus = "restored"
	StatusSold      ProductStatus = "sold"
)

type Product struct {
	ID          uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"` // Unique identifier for the product
	UserID      uuid.UUID     `json:"user_id" gorm:"type:uuid"`                                   // ID of the user who owns the product
	Name        string        `json:"name"`                                                       // Name of the product
	Description string        `json:"description"`                                                // Description of the product
	Price       float64       `json:"price"`                                                      // Price of the product
	SubCategory string        `json:"sub_category"`                                               // Subcategory of the product
	Status      ProductStatus `json:"status" gorm:"type:varchar(20);default:'uncompleted'"`       // Status of the product (uses varchar instead of enum for MySQL)
	Category    string        `json:"category"`                                                   // Category of the product
	CreatedAt   time.Time     `json:"created_at" gorm:"autoCreateTime"`                           // Timestamp when the product was created
}

// ProductResponse represents the structure used to return a product with its associated transactions and user information.
type ProductResponse struct {
	ID           uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"` // Product ID
	User         User          `gorm:"foreignKey:UserID" json:"user"`                              // Associated user (owner of the product)
	Transactions []Transaction `gorm:"foreignKey:ItemID" json:"transactions"`                      // List of transactions related to the product
	Name         string        `json:"name"`                                                       // Product name
	Description  string        `json:"description"`                                                // Product description
	Price        float64       `json:"price"`                                                      // Product price
	SubCategory  string        `json:"sub_category"`                                               // Subcategory of the product
	Category     string        `json:"category"`                                                   // Category of the product
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`                           // Timestamp when the product was created
}

// ProductRequest is used when creating a new product, without including transactions.
type ProductRequest struct {
	UserID      uuid.UUID     `json:"user_id"`                            // ID of the user creating the product
	Name        string        `json:"name"`                               // Name of the product
	Description string        `json:"description"`                        // Description of the product
	Price       float64       `json:"price"`                              // Price of the product
	SubCategory string        `json:"sub_category"`                       // Subcategory of the product
	Category    string        `json:"category"`                           // Category of the product
	Status      ProductStatus `json:"status,omitempty"`                   // Status of the product (optional during request)
	ImageURL    string        `gorm:"type:varchar(255)" json:"image_url"` // URL of the transaction image
}

// Transaction defines the structure for a transaction involving a product.
type Transaction struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey;unique" json:"id"`       // Primary key, unique identifier for each transaction
	ItemID      uuid.UUID         `gorm:"type:uuid;not null" json:"item_id"`           // Reference to the product involved in the transaction
	UserID      uuid.UUID         `gorm:"type:uuid;not null" json:"user_id"`           // Reference to the user performing the transaction
	Description string            `gorm:"type:text" json:"description"`                // Description of the transaction
	Action      TransactionAction `gorm:"type:varchar(20);not null" json:"action"`     // Action type of the transaction
	ImageURL    string            `gorm:"type:varchar(255)" json:"image_url"`          // URL of the transaction image
	CreatedAt   time.Time         `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"` // Transaction timestamp
}

// TransactionAction defines the possible actions for a transaction.
type TransactionAction string

const (
	Submitted   TransactionAction = "submitted"
	Revitalized TransactionAction = "revitalized"
	Sold        TransactionAction = "sold"
)

type TransactionRequest struct {
	ItemID      uuid.UUID         `gorm:"type:uuid;not null" json:"item_id"`       // Reference to the product involved in the transaction
	UserID      uuid.UUID         `gorm:"type:uuid;not null" json:"user_id"`       // Reference to the user performing the transaction
	Description string            `gorm:"type:text" json:"description"`            // Description of the transaction
	Action      TransactionAction `gorm:"type:varchar(20);not null" json:"action"` // Action type of the transaction
	ImageURL    string            `gorm:"type:varchar(255)" json:"image_url"`      // URL of the transaction image
}
type AddTransactionRequest struct {
	Description string            `gorm:"type:text" json:"description"`            // Description of the transaction
	Action      TransactionAction `gorm:"type:varchar(20);not null" json:"action"` // Action type of the transaction
	ImageURL    string            `gorm:"type:varchar(255)" json:"image_url"`      // URL of the transaction image
}
