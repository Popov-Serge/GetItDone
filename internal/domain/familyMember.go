package domain

import (
	"github.com/google/uuid"
	"time"
)

type FamilyMember struct {
	ID             uuid.UUID
	FamilyID       uuid.UUID
	UserID         *uuid.UUID
	Name           string
	RelationTypeID uuid.UUID
	Role           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
