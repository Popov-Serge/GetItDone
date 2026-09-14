package domain

import (
	"time"

	"github.com/google/uuid"
)

type MemberAlias struct {
	ID             uuid.UUID
	FamilyMemberID string
	Alias          string
	CreatedAt      time.Time
}
