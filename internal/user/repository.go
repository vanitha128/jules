package user

import (
	"context"
)

// UserRepository defines the interface for database operations related to users.
// The actual implementation (e.g., for PostgreSQL) will be in internal/database.
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	// ... other methods
}
// Need to import uuid for GetUserByID
import "github.com/google/uuid"

// We will have a struct that implements this interface, likely in another package (e.g., internal/database/postgres_user_repository.go)
// type userRepositoryImpl struct {
// 	db *gorm.DB // or *sql.DB, or whatever DB connection type is used
// }

// func NewUserRepository(db *gorm.DB) UserRepository {
//  return &userRepositoryImpl{db: db}
// }

// func (r *userRepositoryImpl) CreateUser(ctx context.Context, user *User) error {
//  // Actual database insertion logic here
// 	// For example, using GORM:
// 	// if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
// 	// 	return err
// 	// }
// 	// return nil
// 	return errors.New("CreateUser not implemented") // Placeholder
// }
