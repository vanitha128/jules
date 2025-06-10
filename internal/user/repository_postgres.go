package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
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
	// GORM's Create method will set CreatedAt, UpdatedAt if they are time.Time and have `gorm:"autoCreateTime"` etc.
	// It will also generate UUID if configured for the specific dialect or if `gorm:"default:gen_random_uuid()"` is used
	// and the database supports it. Since User.ID is uuid.UUID and primary key, GORM handles it well.
	// We are setting User.ID = uuid.New() in the service layer before calling this.
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		// TODO: Check for specific GORM errors, like unique constraint violations (e.g., email already exists)
		// and return a more specific error.
		return err
	}
	return nil
}

// GetUserByEmail retrieves a user by their email address.
// Returns gorm.ErrRecordNotFound if no user is found.
func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// It's good practice to return a domain-specific error or the error from the `cache` package (ErrNotFound)
			// For now, the service layer handles gorm.ErrRecordNotFound by returning its own ErrInvalidCredentials.
			// Alternatively, we could return a user.ErrNotFound here.
			return nil, err // Let service layer decide how to map this
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID.
// Returns gorm.ErrRecordNotFound if no user is found.
func (r *postgresUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err // Let service layer map this
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user's information in the database.
// GORM's Save method will update all fields if primary key is set, or create if not.
// For partial updates, Updates method is better.
// Given the User object is passed, Save is fine if it's the full object fetched and modified.
// If only specific fields are meant to be updated, ensure only those are in the user struct passed,
// or use `db.Model(&User{}).Where("id = ?", user.ID).Updates(User{FirstName: user.FirstName, ...})`.
// The current service methods (ChangePassword, UpdateProfile) fetch the full user, modify it, then call this.
// So, user.Password or user.FirstName/LastName will be updated.
func (r *postgresUserRepository) UpdateUser(ctx context.Context, user *User) error {
	// user.UpdatedAt will be automatically updated by GORM if `gorm:"autoUpdateTime"` is on the field.
	// Ensure the user object passed has the ID set for GORM to know which record to update.
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// This can happen if Save is called but the record doesn't exist or no fields were changed
		// that GORM considers dirty. If an update was expected, this might indicate an issue.
		// However, GORM's Save won't error if rows affected is 0 but no other error occurred.
		// For an explicit "not found to update" error, you might need to check existence first
		// or rely on service layer logic.
		// For now, we assume if no error, it's fine.
		return nil // Or return a custom error if RowsAffected is 0 and you expect it to be > 0
	}
	return nil
}
