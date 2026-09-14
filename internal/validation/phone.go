package validation

import (
	"strings"

	"github.com/nyaruka/phonenumbers/v2"
)

type PhoneValidator struct{}

func NewPhoneValidator() *PhoneValidator {
	return &PhoneValidator{}
}

func (v *PhoneValidator) Validate(phone string) (string, bool) {
	phone = strings.TrimSpace(phone)

	number, err := phonenumbers.Parse(phone, "RU")

	if err != nil {
		return "", false
	}

	if !phonenumbers.IsValidNumber(number) {
		return "", false
	}

	return phonenumbers.Format(number, phonenumbers.E164), true
}
