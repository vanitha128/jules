package user

import (
	"context"
	"errors"
	"fmt" // Added fmt import
	"time"

	"go-moon/pkg/utils" // Updated import path
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
	jwtUtil  *utils.JWTUtil
}

// NewUserService creates a new UserService.
func NewUserService(userRepo UserRepository, jwtUtil *utils.JWTUtil) UserService {
	return &userService{userRepo: userRepo, jwtUtil: jwtUtil}
}

// Register handles the business logic for user registration.
func (s *userService) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		DOB:       req.DOB,
	}

	err = s.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user in repository: %w", err)
	}
	return newUser, nil
}

// Login handles the business logic for user login. (Corrected version)
func (s *userService) Login(ctx context.Context, req LoginRequest) (string, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token")
	}

	return accessToken, refreshToken, nil
}


// ChangePassword handles changing a user's password.
func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	user.Password = string(hashedNewPassword)
	user.UpdatedAt = time.Now()

	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}

// UpdateProfile handles updating a user's profile information.
func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	updated := false
	if req.FirstName != "" {
		user.FirstName = req.FirstName
		updated = true
	}
	if req.LastName != "" {
		user.LastName = req.LastName
		updated = true
	}

	if !updated {
		return user, nil
	}

	user.UpdatedAt = time.Now()

	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}
	return user, nil
}

// GetProfile retrieves a user's profile information.
func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return user, nil
}

// Helper to create a consistent error for invalid credentials
var ErrInvalidCredentials = errors.New("invalid email or password")
