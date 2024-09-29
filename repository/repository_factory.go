package repository

import (
	"database/sql"
)

// RepositoryFactory defines the interface for creating repositories
type RepositoryFactory interface {
	NewUserRepository() UserRepository
}

// repositoryFactory is the concrete implementation of RepositoryFactory
type repositoryFactory struct {
	db *sql.DB
}

// NewRepositoryFactory creates a new instance of repositoryFactory
func NewRepositoryFactory(db *sql.DB) RepositoryFactory {
	return &repositoryFactory{db: db}
}

// NewUserRepository creates a new instance of UserRepository
func (f *repositoryFactory) NewUserRepository() UserRepository {
	return NewUserRepository(f.db)
}
