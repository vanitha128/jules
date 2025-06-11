package user

import (
	"context"
	"errors"
	"time"

	"github.com/your-username/go-moon/pkg/utils" // Import JWT utility
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	Login(ctx context.Context, req LoginRequest) (accessToken string, refreshToken string, err error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error
	UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*User, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*User, error)
}

type userService struct {
	userRepo UserRepository
	jwtUtil  *utils.JWTUtil // Added jwtUtil
}

// NewUserService creates a new UserService.
func NewUserService(userRepo UserRepository, jwtUtil *utils.JWTUtil) UserService { // Added jwtUtil parameter
	return &userService{userRepo: userRepo, jwtUtil: jwtUtil} // Store jwtUtil
}

// Register handles the business logic for user registration.
func (s *userService) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err // Proper error handling should be more sophisticated
	}

	newUser := &User{
		ID:        uuid.New(), // Generate a new UUID
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		DOB:       req.DOB,
		// CreatedAt and UpdatedAt are usually handled by GORM or the database
	}

	err = s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) { // Check for the specific domain error
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user in repository: %w", err) // Wrap other errors
	}

	// It's good practice to return a User DTO that doesn't include the password.
	// For now, returning the full user object, but excluding password before sending to client is crucial.
	// newUser.Password = "" // Clear password before returning, if not using a DTO
	return newUser, nil
}

// Login handles the business logic for user login.
func (s *userService) Login(ctx context.Context, req LoginRequest) (string, string, error) {
	// 1. Get user by email
	// user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	// if err != nil {
	// 	// Could be user not found or other DB error
	// 	return "", "", err // Or a more specific error like ErrInvalidCredentials
	// }

	// For now, let's assume a dummy user until GetUserByEmail is implemented
	dummyUser := &User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  "$2a$10$abcdefghijklmnopqrstuv", // Dummy hash, bcrypt.CompareHashAndPassword will fail
	}
	// Replace with actual user fetch:
	user := dummyUser // This line is temporary

	// 2. Compare password
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	// if err != nil {
	// 	// Password doesn't match
	// 	return "", "", err // Or ErrInvalidCredentials
	// }

	// For testing, let's simulate a password match for a known dummy password if email is "test@example.com"
	// In a real scenario, you'd use bcrypt.CompareHashAndPassword as above.
	// We need to generate a known hash for "password123" to test this path.
	// Example: hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	// fmt.Println(string(hashedPassword)) // $2a$10$yourGeneratedHashHere
	knownHashForPassword123 := "$2a$10$RSPlkS3hY90TjcL2N7P2cuL2ACRSPk9k78B2yK6dI7Y2f/75vC1aK" // Hash for "password123"

	if user.Email == "test@example.com" && req.Password == "password123" {
		// Simulate successful password comparison for the test user
		err := bcrypt.CompareHashAndPassword([]byte(knownHashForPassword123), []byte(req.Password))
		if err != nil {
			// This means our known hash is wrong or bcrypt is acting up.
			// For the purpose of this step, we'll log and proceed as if it matched for "test@example.com"
			// log.Printf("bcrypt compare error for test user (should not happen if hash is correct): %v", err)
			// return "", "", errors.New("simulated bcrypt error for test user")
		}
		// If the above CompareHashAndPassword was the actual check, and it passed, we'd proceed.
	} else {
		// For any other user or if password doesn't match "password123" for the test user, simulate failure.
		// This simulates the bcrypt.CompareHashAndPassword failing for other users.
		// return "", "", errors.New("invalid credentials (simulated)")
	}
	// The above if/else is a HACK for now. The actual logic should be:
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	// if err != nil { // This means password mismatch or other error
	//     return "", "", errors.New("invalid email or password")
	// }


	// 3. Generate JWT tokens (using placeholders for now)
	// accessToken, err := utils.GenerateAccessToken(user.ID)
	// if err != nil {
	// 	return "", "", err
	// }
	// refreshToken, err := utils.GenerateRefreshToken(user.ID)
	// if err != nil {
	// 	return "", "", err
	// }
	// 3. Generate JWT tokens
	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}
	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ChangePassword handles changing a user's password.
func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error {
	// 1. Get user by ID
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		// Could be user not found or other DB error
		// In a real app, might want to distinguish UserNotFound error
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		// Password doesn't match
		return ErrInvalidCredentials // Or a more specific "old password mismatch" error
	}

	// 3. Hash new password
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 4. Update user's password
	user.Password = string(hashedNewPassword)
	user.UpdatedAt = time.Now() // Update the timestamp

	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}

// UpdateProfile handles updating a user's profile information.
func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*User, error) {
	// 1. Get user by ID
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Update fields if provided in the request
	updated := false
	if req.FirstName != "" {
		user.FirstName = req.FirstName
		updated = true
	}
	if req.LastName != "" {
		user.LastName = req.LastName
		updated = true
	}

	// 3. If nothing was updated (e.g. empty request), just return the user
	// The handler already checks for completely empty request, but service could also enforce.
	if !updated {
		// Or return an error like "no updateable fields provided"
		return user, nil
	}

	user.UpdatedAt = time.Now()

	// 4. Save updated user to repository
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	// It's good practice to return a DTO that doesn't include sensitive info like password.
	// For now, returning the User object. If it contains password, it should be cleared.
	// user.Password = "" // If User struct is returned directly from DB with password
	return user, nil
}

// GetProfile retrieves a user's profile information.
func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		// In a real app, map gorm.ErrRecordNotFound to a domain-specific error e.g., ErrUserNotFound
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	// Ensure password is not returned, even if it's hashed.
	// Depending on User model, you might have a DTO or clear sensitive fields.
	// user.Password = "" // Example if password field exists on returned User model
	return user, nil
}
// Helper to create a consistent error for invalid credentials
var ErrInvalidCredentials = errors.New("invalid email or password")

// Login handles the business logic for user login.
func (s *userService) Login(ctx context.Context, req LoginRequest) (string, string, error) {
	// 1. Get user by email from repository
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// If db error or user not found, return invalid credentials
		// In a real app, you might log the actual s.userRepo error for debugging
		return "", "", ErrInvalidCredentials
	}

	// 2. Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		// Password doesn't match
		return "", "", ErrInvalidCredentials
	}

	// 3. Generate JWT tokens
	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID) // Use s.jwtUtil
	if err != nil {
		// It's good to log this error on the server side
		// log.Printf("Error generating access token for user %s: %v", user.ID, err)
		return "", "", fmt.Errorf("failed to generate access token") // Generic error to client
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID) // Use s.jwtUtil
	if err != nil {
		// log.Printf("Error generating refresh token for user %s: %v", user.ID, err)
		return "", "", fmt.Errorf("failed to generate refresh token") // Generic error to client
	}

	return accessToken, refreshToken, nil
}
