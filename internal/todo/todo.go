package todo

import (
	"time"

	"github.com/google/uuid"
)

// Todo represents a task or item in a to-do list.
type Todo struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	DueDate     time.Time `json:"due_date,omitempty"`
	IsDone      bool      `gorm:"default:false" json:"is_done"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"` // Foreign key to User
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
