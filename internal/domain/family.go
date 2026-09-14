package domain

import (
	"time"

	"github.com/google/uuid"
)

type Family struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name" validate:"required,min=1,max=100"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
