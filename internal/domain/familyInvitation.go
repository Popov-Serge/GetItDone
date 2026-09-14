package domain

import (
	"time"

	"github.com/google/uuid"
)

type FamilyInvitation struct {
	ID             uuid.UUID
	FamilyID       uuid.UUID
	FamilyMemberID uuid.UUID
	Token          string
	ExpiresAt      time.Time
	AcceptedAt     time.Time
	CreatedAt      time.Time
}
