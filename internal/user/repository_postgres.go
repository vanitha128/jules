package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn" // Moved into the import block
	"gorm.io/gorm"
)

// postgresUserRepository implements the UserRepository interface using GORM.
type postgresUserRepository struct {
	db *gorm.DB
}

// NewPostgresUserRepository creates a new instance of postgresUserRepository.
func NewPostgresUserRepository(db *gorm.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

// CreateUser creates a new user record in the database.
func (r *postgresUserRepository) CreateUser(ctx context.Context, user *User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // PostgreSQL unique_violation error code
				return ErrEmailAlreadyExists
			}
		}
		return err
	}
	return nil
}

// GetUserByEmail retrieves a user by their email address.
func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID.
func (r *postgresUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user's information in the database.
func (r *postgresUserRepository) UpdateUser(ctx context.Context, user *User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return nil
	}
	return nil
}
