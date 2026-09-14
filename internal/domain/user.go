package domain

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Phone        string
	Email        string
	PasswordHash string
	Timezone     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
