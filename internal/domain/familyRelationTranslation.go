package domain

import "github.com/google/uuid"

type FamilyRelationTranslation struct {
	RelationTypeID uuid.UUID
	Locale         string
	Name           string
}
