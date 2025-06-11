package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string    `gorm:"not null"`
	FirstName string    `gorm:"type:varchar(50)"`
	LastName  string    `gorm:"type:varchar(50)"`
	DOB       time.Time
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
