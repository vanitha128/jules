package user

import "errors"

// ErrEmailAlreadyExists is returned when trying to register a user with an email that already exists.
var ErrEmailAlreadyExists = errors.New("email already exists")

// ErrUserNotFound is returned when a user is not found.
// (Adding this here as it's a common user-related domain error, might be useful later)
var ErrUserNotFound = errors.New("user not found")
