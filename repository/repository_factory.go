package repository

import "gorm.io/gorm"

// RepositoryFactory holds the database connection and creates repositories
type RepositoryFactory struct {
	db *gorm.DB
}

// NewRepositoryFactory creates a new factory with the GORM database connection
func NewRepositoryFactory(db *gorm.DB) *RepositoryFactory {
	return &RepositoryFactory{db: db}
}

// GetProductRepository returns a new instance of ProductRepository
func (f *RepositoryFactory) GetProductRepository() *ProductRepository {
	return NewProductRepository(f.db)
}

// GetRatingRepository returns a new instance of RatingRepository
func (f *RepositoryFactory) GetRatingRepository() *RatingRepository {
	return NewRatingRepository(f.db)
}

// GetUserRepository returns a new instance of UserRepository
func (f *RepositoryFactory) GetUserRepository() *UserRepository {
	return NewUserRepository(f.db)
}

func (f *RepositoryFactory) GetTransactionRepository() *TransactionRepository {
	return NewTransactionRepository(f.db)
}
