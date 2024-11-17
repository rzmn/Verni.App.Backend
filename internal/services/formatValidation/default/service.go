package defaultFormatValidation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"verni/internal/services/formatValidation"
	"verni/internal/services/logging"
)

func New(logger logging.Service) formatValidation.Service {
	return &defaultService{
		logger: logger,
	}
}

type defaultService struct {
	logger logging.Service
}

func (c *defaultService) ValidateEmailFormat(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email is invalid: %v", err)
	}
	if strings.TrimSpace(email) != email {
		return fmt.Errorf("email is invalid: leading or trailing spaces")
	}
	return nil
}

func (c *defaultService) ValidatePasswordFormat(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password should contain more than 6 characters")
	}
	return nil
}

func (c *defaultService) ValidateDisplayNameFormat(name string) error {
	if !regexp.MustCompile(`^[A-Za-z]+$`).MatchString(name) {
		return fmt.Errorf("display name is invalid: should contain latin characters only")
	}
	if len(name) < 4 {
		return fmt.Errorf("display name is invalid: should contain at least 4 characters")
	}
	return nil
}
