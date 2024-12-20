package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductStatus defines the possible statuses for a product.
type ProductStatus string

const (
	StatusAvailable         ProductStatus = "available"
	StatusRestored          ProductStatus = "restored"
	StatusRestoredAvailable ProductStatus = "restoredAvailable"
	StatusSold              ProductStatus = "sold"
)

// TransactionAction defines the possible actions for a transaction.
type TransactionAction string

const (
	Submitted            TransactionAction = "submitted"
	SubmittedRevitalized TransactionAction = "submittedRevitalized"
	Revitalized          TransactionAction = "revitalized"
	Sold                 TransactionAction = "sold"
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
	ID            uuid.UUID     `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"` // Product ID
	User          User          `gorm:"foreignKey:UserID" json:"user"`                              // Associated user (owner of the product)
	Transactions  []Transaction `gorm:"foreignKey:ItemID" json:"transactions"`                      // List of transactions related to the product
	Name          string        `json:"name"`                                                       // Product name
	Description   string        `json:"description"`                                                // Product description
	Price         float64       `json:"price"`                                                      // Product price
	SubCategory   string        `json:"sub_category"`                                               // Subcategory of the product
	Rating        int           `json:"rating"`                                                     // Product rating
	RatingCount   int           `json:"rating_count"`                                               // Product rating count
	Status        ProductStatus `json:"status,omitempty"`
	RatingAverage float64       `json:"rating_average"`                   // Product rating average
	Category      string        `json:"category"`                         // Category of the product
	CreatedAt     time.Time     `json:"created_at" gorm:"autoCreateTime"` // Timestamp when the product was created
}

// ProductRequest is used when creating a new product, without including transactions.
type ProductRequest struct {
	UserID      uuid.UUID     `json:"user_id"`             // ID of the user creating the product
	Name        string        `json:"name"`                // Name of the product
	Description string        `json:"description"`         // Description of the product
	Price       float64       `json:"price"`               // Price of the product
	SubCategory string        `json:"sub_category"`        // Subcategory of the product
	Category    string        `json:"category"`            // Category of the product
	Status      ProductStatus `json:"status,omitempty"`    // Status of the product (optional during request)
	ImageData   string        `gorm:"-" json:"image_data"` // Base64 encoded image data for the transaction
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

// TransactionRequest defines the fields for creating a transaction with optional image data
type TransactionRequest struct {
	ItemID      uuid.UUID         `gorm:"type:uuid;not null" json:"item_id"`       // Reference to the product involved in the transaction
	UserID      uuid.UUID         `gorm:"type:uuid;not null" json:"user_id"`       // Reference to the user performing the transaction
	Description string            `gorm:"type:text" json:"description"`            // Description of the transaction
	Action      TransactionAction `gorm:"type:varchar(20);not null" json:"action"` // Action type of the transaction
	ImageData   string            `gorm:"-" json:"image_data"`                     // Base64 encoded image data for the transaction
}

// AddTransactionRequest is used to add a transaction with optional image data
type AddTransactionRequest struct {
	Description string            `gorm:"type:text" json:"description"`            // Description of the transaction
	Action      TransactionAction `gorm:"type:varchar(20);not null" json:"action"` // Action type of the transaction
	ImageData   string            `gorm:"-" json:"image_data"`                     // Base64 encoded image data for the transaction
	Price       float64           `gorm:"type:float64" json:"price"`               // Price of the transaction
}

type DetailedProductResponse struct {
	ID            uuid.UUID             `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primary_key"` // Product ID
	User          uuid.UUID             `gorm:"type:uuid;not null" json:"user_id"`                          // Associated user (owner of the product)
	Transactions  []DetailedTransaction `gorm:"foreignKey:ItemID" json:"transactions"`                      // List of transactions related to the product
	Name          string                `json:"name"`                                                       // Product name
	Description   string                `json:"description"`                                                // Product description
	Price         float64               `json:"price"`                                                      // Product price
	SubCategory   string                `json:"sub_category"`                                               // Subcategory of the product
	Rating        int                   `json:"rating"`                                                     // Product rating
	RatingCount   int                   `json:"rating_count"`                                               // Product rating count
	Status        ProductStatus         `json:"status,omitempty"`
	RatingAverage float64               `json:"rating_average"`                   // Product rating average
	Category      string                `json:"category"`                         // Category of the product
	CreatedAt     time.Time             `json:"created_at" gorm:"autoCreateTime"` // Timestamp when the product was created
}

type DetailedTransaction struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey;unique" json:"id"` // Primary key, unique identifier for each transaction
	ItemID      uuid.UUID         `gorm:"type:uuid;not null" json:"item_id"`     // Reference to the product involved in the transaction
	User        User              `gorm:"foreignKey:UserID" json:"user"`
	Description string            `gorm:"type:text" json:"description"` // Description of the transaction
	Action      TransactionAction `gorm:"type:varchar(20);not null" json:"action"`
	ImageURL    string            `gorm:"type:varchar(255)" json:"image_url"` // URL of the transaction image
}
