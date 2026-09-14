package domain

import "github.com/google/uuid"

type FamilyRelationType struct {
	ID        uuid.UUID
	Code      string
	SortOrder int
	IsActive  bool
}
