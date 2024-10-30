package formatValidation

import (
	"fmt"
	"net/mail"
	"strings"
)

type defaultService struct{}

func (s *defaultService) ValidateEmailFormat(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email is invalid: %v", err)
	}
	if strings.TrimSpace(email) != email {
		return fmt.Errorf("email is invalid: leading or trailing spaces")
	}
	return nil
}

func (s *defaultService) ValidatePasswordFormat(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password should contain more than 6 characters")
	}
	return nil
}
